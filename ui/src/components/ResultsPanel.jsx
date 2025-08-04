import React from "react";

function ResultsPanel({ tableData, onAddSelectedToTable, selectedMessages, messages }) {
  const cleanContent = (content) => {
    // Extract content from within pre tags, or use the content as-is if no pre tags
    const preTagMatch = content.match(/<pre>([\s\S]*?)<\/pre>/i);
    if (preTagMatch) {
      return preTagMatch[1].trim();
    }
    // If no pre tags, just trim the content
    return content.trim();
  };

  const handleAddSelected = () => {
    selectedMessages.forEach(index => {
      if (messages[index] && !messages[index].startsWith('You: ')) {
        const cleanedContent = cleanContent(messages[index]);
        if (cleanedContent) {
          onAddSelectedToTable('NodeType', cleanedContent);
        }
      }
    });
  };

  return (
    <div className="bg-white p-4 border-t h-full flex flex-col">
      <div className="flex justify-between items-center mb-2 flex-shrink-0">
        <div className="font-bold text-sm">Results</div>
        {selectedMessages.length > 0 && (
          <button 
            onClick={handleAddSelected}
            className="bg-blue-500 text-white px-3 py-1 rounded text-sm hover:bg-blue-600"
          >
            Add Selected ({selectedMessages.length})
          </button>
        )}
      </div>
      <div className="flex-1 overflow-auto" style={{ minHeight: 0 }}>
        <table className="min-w-full text-sm border">
          <thead>
            <tr className="bg-gray-200">
              <th className="border px-2 py-1">ID</th>
              <th className="border px-2 py-1">NodeType</th>
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
    </div>
  );
}

export default ResultsPanel; 