import React, { useState, useRef, useEffect } from "react";
import ChatPanel from "./components/ChatPanel";
import ToolFlowPanel from "./components/ToolFlowPanel";
import ResultsPanel from "./components/ResultsPanel";
import TopMenu from "./components/TopMenu";
import { SessionManager } from "./utils/sessionManager";
import {renderInterruptInfo} from "./utils/rendering";

// Load environment variables for Vite
const WS_ENDPOINT = import.meta.env.VITE_WS_ENDPOINT || "ws://localhost:8080/ws";

export default function App() {
  const [messages, setMessages] = useState([]);
  const [responses, setResponses] = useState([]); // Track AI responses separately
  const [reasoning, setReasoning] = useState(["Initial reasoning..."]);
  const [loading, setLoading] = useState(false);
  const [selectedMessages, setSelectedMessages] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [sessionInfo, setSessionInfo] = useState(null);
  const ws = useRef(null);

  // Auto-save session data
  useEffect(() => {
    const saveSessionData = () => {
      const sessionData = {
        messages,
        responses, // Include responses in session data
        reasoning,
        tableData,
        selectedMessages
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
      const reasoning = renderInterruptInfo(data.reasoningGraph);
      setMessages(msgs => [...msgs, response]);
      setResponses(responses => [...responses, response]); // Track as AI response
      setReasoning(r => [...r, `Received: ${JSON.stringify(data.reasoningGraph.response_meta)}`]);
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
    setSessionInfo(SessionManager.getSessionInfo());
  };

  return (
    <div className="flex flex-col h-screen overflow-hidden">
      <TopMenu onLoadSession={handleLoadSession} sessionInfo={sessionInfo} />
      <div className="flex flex-1 min-h-0">
        <div className="w-1/2 border-r bg-white flex flex-col min-h-0">
          <ChatPanel 
            onSend={sendMessage} 
            messages={messages} 
            loading={loading}
            onSelectMessage={handleSelectMessage}
            selectedMessages={selectedMessages}
          />
        </div>
        <div className="w-1/2 flex flex-col min-h-0">
          <ToolFlowPanel reasoning={reasoning} />
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
