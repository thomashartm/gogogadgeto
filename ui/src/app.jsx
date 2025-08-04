import React, { useState, useRef, useEffect } from "react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { oneLight } from "react-syntax-highlighter/dist/esm/styles/prism";
import Logo from "./logo";
import { promptPresets } from "./presets";

function formatChatMessage(text) {
  // If the message is wrapped in <pre>...</pre>, render as preformatted text
  const preTagMatch = text.match(/^<pre>([\s\S]*)<\/pre>$/i);
  if (preTagMatch) {
    return <pre style={{whiteSpace: 'pre-wrap', margin: 0, padding: 8, background: '#f5f5f5', borderRadius: 1}}>{preTagMatch[1]}</pre>;
  }
  // Detect code blocks (```lang ... ```)
  const codeBlockRegex = /```([a-zA-Z0-9]*)\n([\s\S]*?)```/g;
  let lastIndex = 0;
  const elements = [];
  let match;
  let key = 0;
  while ((match = codeBlockRegex.exec(text)) !== null) {
    if (match.index > lastIndex) {
      // Text before the code block
      const before = text.slice(lastIndex, match.index);
      elements.push(<span key={key++}>{before.split("\n").map((line, i) => <React.Fragment key={i}>{line}<br /></React.Fragment>)}</span>);
    }
    // Code block
    const lang = match[1] || "text";
    const code = match[2];
    elements.push(
      <SyntaxHighlighter key={key++} language={lang} style={oneLight} customStyle={{margin:0, padding:8, borderRadius:4}}>
        {code}
      </SyntaxHighlighter>
    );
    lastIndex = codeBlockRegex.lastIndex;
  }
  // Remaining text after the last code block
  if (lastIndex < text.length) {
    const rest = text.slice(lastIndex);
    elements.push(<span key={key++}>{rest.split("\n").map((line, i) => <React.Fragment key={i}>{line}<br /></React.Fragment>)}</span>);
  }
  return elements;
}

function ChatPanel({ onSend, messages, loading, onSelectMessage, selectedMessages }) {
  const [input, setInput] = useState("");
  const [selectedPreset, setSelectedPreset] = useState("");
  const chatEndRef = useRef(null);
  useEffect(() => { chatEndRef.current?.scrollIntoView({ behavior: 'smooth' }); }, [messages]);

  const handlePresetSelect = () => {
    const preset = promptPresets.find(p => p.id === selectedPreset);
    if (preset) {
      setInput(preset.prompt);
    }
  };

  const handleMessageClick = (index) => {
    onSelectMessage(index);
  };

  return (
    <div className="flex flex-col h-full">
      <div className="bg-white p-2 border-b font-bold">Conversation</div>
      <div className="flex-1 overflow-y-auto p-2 space-y-2">
        {messages.map((msg, i) => (
          <div 
            key={i} 
            className={`bg-blue-100 rounded p-2 w-fit max-w-[80%] cursor-pointer transition-colors ${
              selectedMessages.includes(i) ? 'ring-2 ring-blue-500 bg-blue-200' : 'hover:bg-blue-150'
            }`}
            onClick={() => handleMessageClick(i)}
          >
            {formatChatMessage(msg)}
          </div>
        ))}
        {loading && (
          <div className="bg-blue-50 rounded p-2 w-fit max-w-[80%] flex items-center">
            <svg className="animate-spin h-5 w-5 text-blue-400" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none"/>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"/>
            </svg>
            <span className="ml-2 text-blue-400">Processing your response...</span>
          </div>
        )}
        <div ref={chatEndRef} />
      </div>
      <form className="flex p-2 border-t bg-white" onSubmit={e => { e.preventDefault(); onSend(input); setInput(""); }}>
        <textarea
          className="flex-1 border rounded px-2 py-1 mr-2 resize-y min-h-[3em]"
          value={input}
          onChange={e => setInput(e.target.value)}
          placeholder="Enter prompt..."
          rows={3}
          style={{ minHeight: '3em', maxHeight: '20em' }}
        />
        <button className="bg-blue-500 text-white px-4 py-1 rounded" type="submit">Send</button>
      </form>
      <div className="p-2 bg-gray-50 border-t">
        <div className="flex items-center gap-2">
          <label className="text-sm font-medium text-gray-700">Quick Presets:</label>
          <select
            className="flex-1 border rounded px-2 py-1 text-sm"
            value={selectedPreset}
            onChange={e => setSelectedPreset(e.target.value)}
          >
            <option value="">Select a preset...</option>
            {promptPresets.map(preset => (
              <option key={preset.id} value={preset.id}>{preset.name}</option>
            ))}
          </select>
          <button
            type="button"
            className="bg-green-500 text-white px-3 py-1 rounded text-sm hover:bg-green-600"
            onClick={handlePresetSelect}
            disabled={!selectedPreset}
          >
            Use Preset
          </button>
        </div>
      </div>
    </div>
  );
}

function ToolFlowPanel({ reasoning }) {
  return (
    <div className="h-full flex flex-col">
      <div className="bg-white p-2 border-b font-bold">Reasoning</div>
      <ul className="p-2 text-xs list-disc list-inside bg-gray-50 flex-1 overflow-y-auto">
        {reasoning.map((r, i) => (
          <li key={i}>{r}</li>
        ))}
      </ul>
    </div>
  );
}

function ResultsPanel({ tableData, onAddSelectedToTable, selectedMessages, messages }) {
  const generateFindingName = (content) => {
    // Extract first line or first few words as finding name
    const firstLine = content.split('\n')[0].trim();
    if (firstLine.length > 0) {
      return firstLine.length > 50 ? firstLine.substring(0, 50) + '...' : firstLine;
    }
    return 'Finding ' + Date.now();
  };

  const handleAddSelected = () => {
    selectedMessages.forEach(index => {
      if (messages[index] && !messages[index].startsWith('You: ')) {
        const findingName = generateFindingName(messages[index]);
        onAddSelectedToTable(findingName, messages[index]);
      }
    });
  };

  return (
    <div className="bg-white p-4 border-t h-full overflow-auto">
      <div className="flex justify-between items-center mb-2">
        <div className="font-bold">Results</div>
        {selectedMessages.length > 0 && (
          <button 
            onClick={handleAddSelected}
            className="bg-blue-500 text-white px-3 py-1 rounded text-sm hover:bg-blue-600"
          >
            Add Selected ({selectedMessages.length})
          </button>
        )}
      </div>
      <table className="min-w-full text-sm border">
        <thead>
          <tr className="bg-gray-200">
            <th className="border px-2 py-1">ID</th>
            <th className="border px-2 py-1">Name</th>
            <th className="border px-2 py-1">Status</th>
            <th className="border px-2 py-1">Content</th>
          </tr>
        </thead>
        <tbody>
          {tableData.length === 0 ? (
            <tr>
              <td colSpan="4" className="border px-2 py-1 text-center text-gray-500">
                No findings added yet. Select messages and click "Add Selected" to add them here.
              </td>
            </tr>
          ) : (
            tableData.map((item, index) => (
              <tr key={index}>
                <td className="border px-2 py-1">{item.id}</td>
                <td className="border px-2 py-1">{item.name}</td>
                <td className="border px-2 py-1">{item.status}</td>
                <td className="border px-2 py-1 max-w-xs truncate" title={item.content}>
                  {item.content.length > 100 ? item.content.substring(0, 100) + '...' : item.content}
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}

function TopMenu() {
  return (
    <div className="flex items-center justify-between w-full h-12 bg-gray-800 text-white px-4 border-b shadow-sm">
      <div className="flex items-center space-x-2 ">
        <Logo width={28} height={28} />
        <span className="font-bold text-lg tracking-wide">gogogadgeto <span className="font-normal">scanner</span></span>
      </div>
      <div className="flex items-center space-x-2">
        <button className="bg-gray-700 hover:bg-gray-600 px-3 py-1 rounded text-sm">Button 1</button>
        <button className="bg-gray-700 hover:bg-gray-600 px-3 py-1 rounded text-sm">Button 2</button>
        <button className="bg-gray-700 hover:bg-gray-600 px-3 py-1 rounded text-sm">Button 3</button>
        <button className="bg-gray-700 hover:bg-gray-600 px-3 py-1 rounded text-sm">Button 4</button>
        <button className="bg-gray-700 hover:bg-gray-600 px-3 py-1 rounded text-sm">Button 5</button>
      </div>
    </div>
  );
}

export default function App() {
  const [messages, setMessages] = useState([]);
  const [reasoning, setReasoning] = useState(["Initial reasoning..."]);
  const [loading, setLoading] = useState(false);
  const [selectedMessages, setSelectedMessages] = useState([]);
  const [tableData, setTableData] = useState([]);
  const ws = useRef(null);

  useEffect(() => {
    ws.current = new WebSocket("ws://localhost:8080/ws");
    ws.current.onmessage = (event) => {
      setMessages(msgs => [...msgs, event.data]);
      setReasoning(r => [...r, `Received: ${event.data}`]);
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

  return (
    <div className="flex flex-col h-screen">
      <TopMenu />
      <div className="flex flex-1">
        <div className="w-1/2 h-full border-r bg-white flex flex-col">
          <ChatPanel 
            onSend={sendMessage} 
            messages={messages} 
            loading={loading}
            onSelectMessage={handleSelectMessage}
            selectedMessages={selectedMessages}
          />
        </div>
        <div className="w-1/2 h-full flex flex-col">
          <ToolFlowPanel reasoning={reasoning} />
        </div>
      </div>
      <div className="h-48">
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
