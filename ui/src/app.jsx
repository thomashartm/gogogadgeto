import React, { useState, useRef, useEffect } from "react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { oneLight } from "react-syntax-highlighter/dist/esm/styles/prism";
import Logo from "./logo";

function formatChatMessage(text) {
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

function ChatPanel({ onSend, messages }) {
  const [input, setInput] = useState("");
  const chatEndRef = useRef(null);
  useEffect(() => { chatEndRef.current?.scrollIntoView({ behavior: 'smooth' }); }, [messages]);
  return (
    <div className="flex flex-col h-full">
      <div className="bg-white p-2 border-b font-bold">Conversation</div>
      <div className="flex-1 overflow-y-auto p-2 space-y-2">
        {messages.map((msg, i) => (
          <div key={i} className="bg-blue-100 rounded p-2 w-fit max-w-[80%]">{formatChatMessage(msg)}</div>
        ))}
        <div ref={chatEndRef} />
      </div>
      <form className="flex p-2 border-t bg-white" onSubmit={e => { e.preventDefault(); onSend(input); setInput(""); }}>
        <input className="flex-1 border rounded px-2 py-1 mr-2" value={input} onChange={e => setInput(e.target.value)} placeholder="Enter message..." />
        <button className="bg-blue-500 text-white px-4 py-1 rounded" type="submit">Send</button>
      </form>
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

function ResultsPanel() {
  return (
    <div className="bg-white p-4 border-t h-full overflow-auto">
      <div className="font-bold mb-2">Results</div>
      <table className="min-w-full text-sm border">
        <thead>
          <tr className="bg-gray-200">
            <th className="border px-2 py-1">ID</th>
            <th className="border px-2 py-1">Name</th>
            <th className="border px-2 py-1">Status</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td className="border px-2 py-1">1</td>
            <td className="border px-2 py-1">Dummy</td>
            <td className="border px-2 py-1">OK</td>
          </tr>
          <tr>
            <td className="border px-2 py-1">2</td>
            <td className="border px-2 py-1">Test</td>
            <td className="border px-2 py-1">Pending</td>
          </tr>
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
  const ws = useRef(null);

  useEffect(() => {
    ws.current = new WebSocket("ws://localhost:8080/ws");
    ws.current.onmessage = (event) => {
      setMessages(msgs => [...msgs, event.data]);
      setReasoning(r => [...r, `Received: ${event.data}`]);
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
    } else {
      setReasoning(r => [...r, "WebSocket not connected"]);
    }
  };

  return (
    <div className="flex flex-col h-screen">
      <TopMenu />
      <div className="flex flex-1">
        <div className="w-1/2 h-full border-r bg-white flex flex-col">
          <ChatPanel onSend={sendMessage} messages={messages} />
        </div>
        <div className="w-1/2 h-full flex flex-col">
          <ToolFlowPanel reasoning={reasoning} />
        </div>
      </div>
      <div className="h-48">
        <ResultsPanel />
      </div>
    </div>
  );
}
