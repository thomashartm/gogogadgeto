import React from "react";
import Logo from "../logo";
import SessionControls from "./SessionControls";

function TopMenu({ onLoadSession, sessionInfo }) {
  return (
    <div className="flex items-center justify-between w-full h-12 bg-gray-800 text-white px-4 border-b shadow-sm">
      <div className="flex items-center space-x-2 ">
        <Logo width={28} height={28} />
        <span className="font-bold text-lg tracking-wide">GogogGadgeto <span className="font-normal">Security</span></span>
      </div>
      <div className="flex items-center space-x-2">
        <SessionControls onLoadSession={onLoadSession} sessionInfo={sessionInfo} />
        <button className="bg-gray-700 hover:bg-gray-600 px-3 py-1 rounded text-sm">Tools</button>
      </div>
    </div>
  );
}

export default TopMenu; 