// Session Manager for Local Storage persistence and Backend Session Management
export class SessionManager {
  static SESSION_KEY = 'gogogadgeto_session';
  static BACKEND_SESSION_KEY = 'gogogadgeto_backend_session_id';
  static AUTO_SAVE_INTERVAL = 30000; // 30 seconds
  static API_BASE = 'http://localhost:8080/api/session';

  static saveSession(sessionData) {
    try {
      const session = {
        timestamp: Date.now(),
        version: '1.0',
        data: {
          messages: sessionData.messages || [],
          reasoning: sessionData.reasoning || [],
          tableData: sessionData.tableData || [],
          selectedMessages: sessionData.selectedMessages || [],
          responses: sessionData.responses || [] // Store AI responses separately
        }
      };
      
      localStorage.setItem(this.SESSION_KEY, JSON.stringify(session));
      console.log('Session saved successfully');
      return true;
    } catch (error) {
      console.error('Failed to save session:', error);
      return false;
    }
  }

  static loadSession() {
    try {
      const sessionString = localStorage.getItem(this.SESSION_KEY);
      if (!sessionString) {
        return null;
      }

      const session = JSON.parse(sessionString);
      
      // Validate session structure
      if (!session.data || !session.timestamp) {
        console.warn('Invalid session format, clearing corrupted data');
        this.clearSession();
        return null;
      }

      console.log('Session loaded successfully');
      return session.data;
    } catch (error) {
      console.error('Failed to load session:', error);
      this.clearSession();
      return null;
    }
  }

  static clearSession() {
    try {
      localStorage.removeItem(this.SESSION_KEY);
      console.log('Session cleared successfully');
      return true;
    } catch (error) {
      console.error('Failed to clear session:', error);
      return false;
    }
  }

  static getSessionInfo() {
    try {
      const sessionString = localStorage.getItem(this.SESSION_KEY);
      if (!sessionString) {
        return null;
      }

      const session = JSON.parse(sessionString);
      return {
        timestamp: session.timestamp,
        messageCount: session.data?.messages?.length || 0,
        tableDataCount: session.data?.tableData?.length || 0,
        responseCount: session.data?.responses?.length || 0, // Track response count
        size: new Blob([sessionString]).size
      };
    } catch (error) {
      console.error('Failed to get session info:', error);
      return null;
    }
  }

  static exportSession() {
    try {
      const sessionString = localStorage.getItem(this.SESSION_KEY);
      if (!sessionString) {
        return null;
      }

      const session = JSON.parse(sessionString);
      const blob = new Blob([JSON.stringify(session, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      
      const a = document.createElement('a');
      a.href = url;
      a.download = `gogogadgeto-session-${new Date(session.timestamp).toISOString().split('T')[0]}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      
      return true;
    } catch (error) {
      console.error('Failed to export session:', error);
      return false;
    }
  }

  static importSession(file) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      
      reader.onload = (e) => {
        try {
          const session = JSON.parse(e.target.result);
          
          // Validate imported session
          if (!session.data || !session.timestamp) {
            throw new Error('Invalid session format');
          }

          // Save imported session
          localStorage.setItem(this.SESSION_KEY, JSON.stringify(session));
          console.log('Session imported successfully');
          resolve(session.data);
        } catch (error) {
          console.error('Failed to import session:', error);
          reject(error);
        }
      };

      reader.onerror = () => {
        reject(new Error('Failed to read file'));
      };

      reader.readAsText(file);
    });
  }

  // Backend session management methods
  static async createBackendSession() {
    try {
      const response = await fetch(`${this.API_BASE}/new`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const session = await response.json();
      
      // Store session ID in local storage
      localStorage.setItem(this.BACKEND_SESSION_KEY, session.sessionId);
      console.log('Backend session created:', session.sessionId);
      
      return session;
    } catch (error) {
      console.error('Failed to create backend session:', error);
      return null;
    }
  }

  static getBackendSessionId() {
    return localStorage.getItem(this.BACKEND_SESSION_KEY);
  }

  static async sendMessageToBackend(message, sessionId = null) {
    try {
      const currentSessionId = sessionId || this.getBackendSessionId();
      
      const response = await fetch(`${this.API_BASE}/message`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          sessionId: currentSessionId,
          message: message
        })
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      
      // Update stored session ID if it changed (e.g., new session was created)
      if (result.sessionId && result.sessionId !== currentSessionId) {
        localStorage.setItem(this.BACKEND_SESSION_KEY, result.sessionId);
      }
      
      return result;
    } catch (error) {
      console.error('Failed to send message to backend:', error);
      return null;
    }
  }

  static async getBackendSessionHistory(sessionId = null) {
    try {
      const currentSessionId = sessionId || this.getBackendSessionId();
      if (!currentSessionId) {
        return null;
      }

      const response = await fetch(`${this.API_BASE}/${currentSessionId}/history`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Failed to get backend session history:', error);
      return null;
    }
  }

  static async clearBackendSession(sessionId = null) {
    try {
      const currentSessionId = sessionId || this.getBackendSessionId();
      if (!currentSessionId) {
        return true;
      }

      const response = await fetch(`${this.API_BASE}/${currentSessionId}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      // Clear local storage
      localStorage.removeItem(this.BACKEND_SESSION_KEY);
      console.log('Backend session cleared');
      
      return true;
    } catch (error) {
      console.error('Failed to clear backend session:', error);
      return false;
    }
  }

  static saveSessionData(sessionData) {
    // Enhanced session data saving that includes backend session ID
    try {
      const session = {
        timestamp: Date.now(),
        version: '1.0',
        backendSessionId: this.getBackendSessionId(),
        data: {
          messages: sessionData.messages || [],
          reasoning: sessionData.reasoning || [],
          tableData: sessionData.tableData || [],
          selectedMessages: sessionData.selectedMessages || [],
          responses: sessionData.responses || []
        }
      };
      
      localStorage.setItem(this.SESSION_KEY, JSON.stringify(session));
      console.log('Enhanced session saved successfully with backend session ID');
      return true;
    } catch (error) {
      console.error('Failed to save enhanced session:', error);
      return false;
    }
  }
} 