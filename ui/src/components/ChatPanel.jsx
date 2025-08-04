import React, { useState, useRef, useEffect } from "react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { oneLight } from "react-syntax-highlighter/dist/esm/styles/prism";
import { promptPresets } from "../presets";

function formatChatMessage(text) {
  // If the message is wrapped in <pre>...</pre>, render as preformatted text
  const preTagMatch = text.match(/^<pre>([\s\S]*?)<\/pre>$/i);
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
      <div className="bg-white p-2 border-b font-bold flex-shrink-0 text-sm">Conversation</div>
      <div className="flex-1 overflow-y-auto p-2 space-y-2" style={{ minHeight: 0 }}>
        {messages.map((msg, i) => (
          <div 
            key={i} 
            className={`bg-blue-100 rounded p-2 w-fit max-w-[80%] cursor-pointer transition-colors text-sm ${
              selectedMessages.includes(i) ? 'ring-2 ring-blue-500 bg-blue-200' : 'hover:bg-blue-150'
            }`}
            onClick={() => handleMessageClick(i)}
          >
            {formatChatMessage(msg)}
          </div>
        ))}
        {loading && (
          <div className="bg-blue-50 rounded p-2 w-fit max-w-[80%] flex items-center text-sm">
            <svg className="animate-spin h-4 w-4 text-blue-400" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none"/>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"/>
            </svg>
            <span className="ml-2 text-blue-400">Processing your response...</span>
          </div>
        )}
        <div ref={chatEndRef} />
      </div>
      <div className="flex-shrink-0">
        <form className="flex p-2 border-t bg-white" onSubmit={e => { e.preventDefault(); onSend(input); setInput(""); }}>
          <textarea
            className="flex-1 border rounded px-2 py-1 mr-2 resize-y min-h-[3em] text-sm"
            value={input}
            onChange={e => setInput(e.target.value)}
            placeholder="Enter prompt..."
            rows={3}
            style={{ minHeight: '3em', maxHeight: '20em' }}
          />
          <button className="bg-blue-500 text-white px-4 py-1 rounded text-sm" type="submit">Send</button>
        </form>
        <div className="p-2 bg-gray-50 border-t">
          <div className="flex items-center gap-2">
            <label className="text-xs font-medium text-gray-700">Quick Presets:</label>
            <select
              className="flex-1 border rounded px-2 py-1 text-xs"
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
              className="bg-green-500 text-white px-3 py-1 rounded text-xs hover:bg-green-600"
              onClick={handlePresetSelect}
              disabled={!selectedPreset}
            >
              Use Preset
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default ChatPanel; 