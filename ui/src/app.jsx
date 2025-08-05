import React, { useState, useRef, useEffect } from "react";
import ChatPanel from "./components/ChatPanel";
import ToolFlowPanel from "./components/ToolFlowPanel";
import ResultsPanel from "./components/ResultsPanel";
import TopMenu from "./components/TopMenu";
import ResizableDivider from "./components/ResizableDivider";
import { SessionManager } from "./utils/sessionManager";
import {renderInterruptInfo} from "./utils/rendering";

// Load environment variables for Vite
const WS_ENDPOINT = import.meta.env.VITE_WS_ENDPOINT || "ws://localhost:8080/ws";

// Helper function to redact content fields in history
function redactHistoryFields(obj) {
  if (obj === null || obj === undefined) return obj;
  
  if (Array.isArray(obj)) {
    return obj.map(item => redactHistoryFields(item));
  }
  
  if (typeof obj === 'object') {
    const newObj = {};
    for (const key in obj) {
      if (key === 'content') {
        // Replace content with boolean indicator
        newObj[key] = obj[key] && obj[key].trim().length > 0;
      } else {
        newObj[key] = redactHistoryFields(obj[key]);
      }
    }
    return newObj;
  }
  
  return obj;
}

export default function App() {
  const [messages, setMessages] = useState([]);
  const [responses, setResponses] = useState([]); // Track AI responses separately
  const [reasoning, setReasoning] = useState(["Initial reasoning..."]);
  const [loading, setLoading] = useState(false);
  const [selectedMessages, setSelectedMessages] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [sessionInfo, setSessionInfo] = useState(null);
  const [leftPanelWidth, setLeftPanelWidth] = useState(50); // Default 50% for chat panel
  const ws = useRef(null);

  // Auto-save session data
  useEffect(() => {
    const saveSessionData = () => {
      const sessionData = {
        messages,
        responses, // Include responses in session data
        reasoning,
        tableData,
        selectedMessages,
        leftPanelWidth
      };
      
      SessionManager.saveSession(sessionData);
      setSessionInfo(SessionManager.getSessionInfo());
    };

    // Auto-save every 30 seconds
    const interval = setInterval(saveSessionData, SessionManager.AUTO_SAVE_INTERVAL);

    // Also save when component unmounts
    return () => {
      clearInterval(interval);
      saveSessionData();
    };
  }, [messages, responses, reasoning, tableData, selectedMessages]);

  // Auto-restore session on app load
  useEffect(() => {
    const restoreSession = () => {
      const sessionData = SessionManager.loadSession();
      if (sessionData) {
        // Check if user wants to restore the session
        const lastSessionInfo = SessionManager.getSessionInfo();
        if (lastSessionInfo && (lastSessionInfo.messageCount > 0 || lastSessionInfo.responseCount > 0)) {
          const shouldRestore = window.confirm(
            `Found a previous session with ${lastSessionInfo.messageCount} messages, ${lastSessionInfo.responseCount} responses, and ${lastSessionInfo.tableDataCount} table items. Would you like to restore it?`
          );
          
          if (shouldRestore) {
            setMessages(sessionData.messages || []);
            setResponses(sessionData.responses || []); // Restore responses
            setReasoning(sessionData.reasoning || ["Initial reasoning..."]);
            setTableData(sessionData.tableData || []);
            setSelectedMessages(sessionData.selectedMessages || []);
            console.log('Session restored successfully');
          }
        }
      }
      
      // Update session info
      setSessionInfo(SessionManager.getSessionInfo());
    };

    restoreSession();
  }, []);

  useEffect(() => {
    ws.current = new WebSocket(WS_ENDPOINT);
    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      console.log("WebSocket message received:", data);
      const response = data.response;
      setMessages(msgs => [...msgs, response]);
      setResponses(responses => [...responses, response]); // Track as AI response
      
      // Keep reasoning as original JSON, only redact history content fields
      setReasoning(r => [...r, `Reasoning: ${JSON.stringify(data.reasoning)}`]);
      
      const historyWithRedactedFields = redactHistoryFields(data.history);
      setReasoning(r => [...r, `History: ${JSON.stringify(historyWithRedactedFields)}`]);
      
      setLoading(false);
    };
    ws.current.onclose = () => setReasoning(r => [...r, "WebSocket disconnected"]);
    ws.current.onerror = (e) => setReasoning(r => [...r, "WebSocket error"]);
    return () => ws.current && ws.current.close();
  }, []);

  const sendMessage = (msg) => {
    if (ws.current && ws.current.readyState === 1) {
      ws.current.send(msg);
      setMessages(msgs => [...msgs, `You: ${msg}`]);
      setReasoning(r => [...r, `Sent: ${msg}`]);
      setLoading(true);
    } else {
      setReasoning(r => [...r, "WebSocket not connected"]);
    }
  };

  const handleSelectMessage = (index) => {
    setSelectedMessages(prev => {
      if (prev.includes(index)) {
        return prev.filter(i => i !== index);
      } else {
        return [...prev, index];
      }
    });
  };

  const handleAddSelectedToTable = (name, content) => {
    const newItem = {
      id: tableData.length + 1,
      name: name,
      status: 'Active',
      content: content
    };
    setTableData(prev => [...prev, newItem]);
    setSelectedMessages([]); // Clear selection after adding
  };

  const handleLoadSession = (sessionData) => {
    setMessages(sessionData.messages || []);
    setResponses(sessionData.responses || []); // Restore responses
    setReasoning(sessionData.reasoning || ["Initial reasoning..."]);
    setTableData(sessionData.tableData || []);
    setSelectedMessages(sessionData.selectedMessages || []);
    setLeftPanelWidth(sessionData.leftPanelWidth || 50); // Restore panel width, default to 50%
    setSessionInfo(SessionManager.getSessionInfo());
  };

  const handleClearReasoning = () => {
    setReasoning([]);
  };

  const handlePanelResize = (newLeftWidth) => {
    setLeftPanelWidth(newLeftWidth);
  };

  return (
    <div className="flex flex-col h-screen overflow-hidden">
      <TopMenu onLoadSession={handleLoadSession} sessionInfo={sessionInfo} />
      <div className="flex flex-1 min-h-0 resizable-container">
        <div 
          className="bg-white flex flex-col min-h-0 border-r"
          style={{ width: `${leftPanelWidth}%` }}
        >
          <ChatPanel 
            onSend={sendMessage} 
            messages={messages} 
            loading={loading}
            onSelectMessage={handleSelectMessage}
            selectedMessages={selectedMessages}
          />
        </div>
        <ResizableDivider onResize={handlePanelResize} />
        <div 
          className="flex flex-col min-h-0"
          style={{ width: `${100 - leftPanelWidth}%` }}
        >
          <ToolFlowPanel reasoning={reasoning} onClear={handleClearReasoning} />
        </div>
      </div>
      <div className="h-48 flex-shrink-0">
        <ResultsPanel 
          tableData={tableData}
          onAddSelectedToTable={handleAddSelectedToTable}
          selectedMessages={selectedMessages}
          messages={messages}
        />
      </div>
    </div>
  );
}
