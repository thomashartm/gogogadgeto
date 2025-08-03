import React from "react";

// Inspector Gadget homage logo
export default function Logo({ width = 28, height = 28 }) {
  return (
    <svg width={width} height={height} viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="16" cy="16" r="16" fill="#4F8EF7" />
      <ellipse cx="16" cy="20" rx="7" ry="5" fill="#fff" />
      <ellipse cx="16" cy="13" rx="6" ry="7" fill="#fff" />
      <ellipse cx="16" cy="13" rx="4" ry="5" fill="#4F8EF7" />
      <rect x="14" y="3" width="4" height="7" rx="2" fill="#fff" stroke="#4F8EF7" strokeWidth="1.5" />
      <rect x="7" y="8" width="3" height="4" rx="1.5" fill="#fff" stroke="#4F8EF7" strokeWidth="1.2" />
      <rect x="22" y="8" width="3" height="4" rx="1.5" fill="#fff" stroke="#4F8EF7" strokeWidth="1.2" />
    </svg>
  );
}
