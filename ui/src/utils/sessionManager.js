// Session Manager for Local Storage persistence
export class SessionManager {
  static SESSION_KEY = 'gogogadgeto_session';
  static AUTO_SAVE_INTERVAL = 30000; // 30 seconds

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
} 