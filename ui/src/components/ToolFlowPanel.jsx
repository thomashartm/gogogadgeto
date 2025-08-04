import React from "react";

function ToolFlowPanel({ reasoning }) {
  return (
    <div className="h-full flex flex-col">
      <div className="bg-white p-2 border-b font-bold flex-shrink-0 text-sm">Reasoning</div>
      <ul className="p-2 text-xs list-disc list-inside bg-gray-50 flex-1 overflow-y-auto" style={{ minHeight: 0 }}>
        {reasoning.map((r, i) => (
          <li key={i}>{r}</li>
        ))}
      </ul>
    </div>
  );
}

export default ToolFlowPanel; 