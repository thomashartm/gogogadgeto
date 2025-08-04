import React, { useState, useRef } from "react";
import { SessionManager } from "../utils/sessionManager";

function SessionControls({ onLoadSession, sessionInfo }) {
  const [showImportDialog, setShowImportDialog] = useState(false);
  const [importError, setImportError] = useState("");
  const fileInputRef = useRef(null);

  const handleSaveSession = () => {
    const success = SessionManager.saveSession(sessionInfo);
    if (success) {
      // You could add a toast notification here
      console.log('Session saved manually');
    }
  };

  const handleLoadSession = () => {
    const sessionData = SessionManager.loadSession();
    if (sessionData) {
      onLoadSession(sessionData);
    }
  };

  const handleClearSession = () => {
    if (window.confirm('Are you sure you want to clear the current session? This action cannot be undone.')) {
      SessionManager.clearSession();
      // Reload the page to reset all state
      window.location.reload();
    }
  };

  const handleExportSession = () => {
    const success = SessionManager.exportSession();
    if (!success) {
      alert('Failed to export session. Please try again.');
    }
  };

  const handleImportSession = () => {
    setShowImportDialog(true);
    setImportError("");
  };

  const handleFileSelect = (event) => {
    const file = event.target.files[0];
    if (!file) return;

    if (file.type !== 'application/json' && !file.name.endsWith('.json')) {
      setImportError('Please select a valid JSON file');
      return;
    }

    SessionManager.importSession(file)
      .then((sessionData) => {
        onLoadSession(sessionData);
        setShowImportDialog(false);
        setImportError("");
        if (fileInputRef.current) {
          fileInputRef.current.value = '';
        }
      })
      .catch((error) => {
        setImportError(`Import failed: ${error.message}`);
      });
  };

  const formatFileSize = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatTimestamp = (timestamp) => {
    return new Date(timestamp).toLocaleString();
  };

  return (
    <div className="flex items-center space-x-2">
      {/* Session Info Display */}
      {sessionInfo && (
        <div className="text-xs text-gray-300 mr-2">
          <span title={`Last saved: ${formatTimestamp(sessionInfo.timestamp)}`}>
            💾 {sessionInfo.messageCount}msgs, {sessionInfo.responseCount}resp, {sessionInfo.tableDataCount}items
          </span>
          <span className="ml-1" title="Session size">
            ({formatFileSize(sessionInfo.size)})
          </span>
        </div>
      )}

      {/* Session Control Buttons */}
      <button 
        onClick={handleSaveSession}
        className="bg-green-600 hover:bg-green-700 px-2 py-1 rounded text-xs"
        title="Save current session"
      >
        💾 Save
      </button>
      
      <button 
        onClick={handleLoadSession}
        className="bg-blue-600 hover:bg-blue-700 px-2 py-1 rounded text-xs"
        title="Load last saved session"
      >
        📂 Load
      </button>
      
      <button 
        onClick={handleExportSession}
        className="bg-purple-600 hover:bg-purple-700 px-2 py-1 rounded text-xs"
        title="Export session to file"
      >
        📤 Export
      </button>
      
      <button 
        onClick={handleImportSession}
        className="bg-orange-600 hover:bg-orange-700 px-2 py-1 rounded text-xs"
        title="Import session from file"
      >
        📥 Import
      </button>
      
      <button 
        onClick={handleClearSession}
        className="bg-red-600 hover:bg-red-700 px-2 py-1 rounded text-xs"
        title="Clear current session"
      >
        🗑️ Clear
      </button>

      {/* Import Dialog */}
      {showImportDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-4 rounded-lg max-w-md w-full mx-4">
            <h3 className="text-lg font-bold mb-2">Import Session</h3>
            <p className="text-sm text-gray-600 mb-4">
              Select a JSON file to import a previously exported session.
            </p>
            
            <input
              ref={fileInputRef}
              type="file"
              accept=".json"
              onChange={handleFileSelect}
              className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
            />
            
            {importError && (
              <p className="text-red-500 text-sm mt-2">{importError}</p>
            )}
            
            <div className="flex justify-end space-x-2 mt-4">
              <button
                onClick={() => setShowImportDialog(false)}
                className="px-3 py-1 text-sm bg-gray-300 hover:bg-gray-400 rounded"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default SessionControls; 