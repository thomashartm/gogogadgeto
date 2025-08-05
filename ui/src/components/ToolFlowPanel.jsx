import React from "react";

function ToolFlowPanel({ reasoning, onClear }) {
  // Parse reasoning entries to extract key-value pairs
  const parseEntry = (entry) => {
    // Check if entry has key-value format like "Key: value"
    const colonIndex = entry.indexOf(": ");
    if (colonIndex > 0 && colonIndex < 50) { // Reasonable key length
      const key = entry.substring(0, colonIndex);
      const value = entry.substring(colonIndex + 2);
      return { key, value };
    }
    // If no clear key-value format, treat the whole thing as value
    return { key: null, value: entry };
  };

  const formatValue = (value) => {
    // Try to format JSON strings nicely
    try {
      const parsed = JSON.parse(value);
      return JSON.stringify(parsed, null, 2);
    } catch {
      return value;
    }
  };

  return (
    <div className="h-full flex flex-col bg-white">
      {/* Header with title and clear button */}
      <div className="bg-gray-100 p-1 border-b flex justify-between items-center flex-shrink-0">
        <h3 className="font-bold text-sm text-gray-800">Tool Flow & Reasoning</h3>
        <button
          onClick={onClear}
          className="px-3 py-1 text-xs bg-red-500 text-white rounded hover:bg-red-600 transition-colors"
          title="Clear tool flow panel"
        >
          Clear
        </button>
      </div>
      
      {/* Scrollable content area */}
      <div className="flex-1 overflow-y-auto" style={{ minHeight: 0 }}>
        <div className="p-1 space-y-1">
          {reasoning.map((entry, i) => {
            const { key, value } = parseEntry(entry);
            
            return (
              <div 
                key={i} 
                className="w-full bg-gray-50 border border-gray-200 rounded p-1 text-xs font-mono"
              >
                {key ? (
                  <div className="space-y-1">
                    <div className="font-bold text-gray-700 text-xs uppercase tracking-wide">
                      {key}
                    </div>
                    <div className="text-gray-600 whitespace-pre-wrap break-words">
                      {formatValue(value)}
                    </div>
                  </div>
                ) : (
                  <div className="text-gray-600 whitespace-pre-wrap break-words">
                    {value}
                  </div>
                )}
              </div>
            );
          })}
          
          {reasoning.length === 0 && (
            <div className="text-center text-gray-400 text-xs italic py-8">
              No tool flow data yet...
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default ToolFlowPanel; 