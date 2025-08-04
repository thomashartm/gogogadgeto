import React from "react";

export function renderInterruptInfo({ info }) {
    return (
        <code
            style={{
                whiteSpace: "pre",
                overflowX: "auto",
                display: "block",
                background: "#f5f5f5",
                padding: "8px",
                borderRadius: "4px",
                fontFamily: "monospace"
            }}
        >
            {JSON.stringify(info)}
        </code>
    );
}