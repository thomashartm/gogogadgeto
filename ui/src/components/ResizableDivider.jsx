import React, { useState, useCallback } from 'react';

const ResizableDivider = ({ onResize, className = '' }) => {
  const [isDragging, setIsDragging] = useState(false);

  const handleMouseDown = useCallback((e) => {
    e.preventDefault();
    setIsDragging(true);

    const handleMouseMove = (e) => {
      const containerRect = e.currentTarget.getBoundingClientRect ? 
        e.currentTarget.getBoundingClientRect() : 
        document.querySelector('.resizable-container').getBoundingClientRect();
      
      const newWidth = ((e.clientX - containerRect.left) / containerRect.width) * 100;
      
      // Constrain between 20% and 80%
      const constrainedWidth = Math.max(20, Math.min(80, newWidth));
      onResize(constrainedWidth);
    };

    const handleMouseUp = () => {
      setIsDragging(false);
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
  }, [onResize]);

  return (
    <div
      className={`
        w-1 bg-gray-300 hover:bg-blue-400 cursor-col-resize transition-colors duration-150 
        ${isDragging ? 'bg-blue-500' : ''} 
        ${className}
      `}
      onMouseDown={handleMouseDown}
      style={{
        userSelect: 'none',
        minWidth: '4px',
        maxWidth: '4px'
      }}
    >
      {/* Visual indicator dots */}
      <div className="h-full flex flex-col justify-center items-center opacity-50">
        <div className="w-0.5 h-1 bg-gray-600 mb-1 rounded-full"></div>
        <div className="w-0.5 h-1 bg-gray-600 mb-1 rounded-full"></div>
        <div className="w-0.5 h-1 bg-gray-600 mb-1 rounded-full"></div>
        <div className="w-0.5 h-1 bg-gray-600 mb-1 rounded-full"></div>
        <div className="w-0.5 h-1 bg-gray-600 rounded-full"></div>
      </div>
    </div>
  );
};

export default ResizableDivider; 