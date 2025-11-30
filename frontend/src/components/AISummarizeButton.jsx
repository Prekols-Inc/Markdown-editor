import { useState } from "react";
import { BrainCircuit } from "lucide-react";
import { summarizeWithGigachat } from "../gigachat";

export default function AISummarizeButton({ current }) {
    const [loading, setLoading] = useState(false);
    const [summary, setSummary] = useState(null);
    const [showModal, setShowModal] = useState(false);

    const handleSummarize = async () => {
        if (!current) return;
        setLoading(true);
        try {
            const result = await summarizeWithGigachat(current.text);
            setSummary(result);
            setShowModal(true);
        } catch (e) {
            console.error("AI error:", e);
            setSummary("Ошибка при запросе AI.");
            setShowModal(true);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="ai-summarize-container" style={{ padding: "10px" }}>
            <button
                className="btn"
                style={{
                    width: "100%",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    gap: 8,
                    opacity: loading ? 0.5 : 1,
                    animation: loading ? "pulse 1.2s infinite" : "none"
                }}
                disabled={loading || !current}
                onClick={handleSummarize}
            >
                <BrainCircuit size={20} strokeWidth={1.75} />
                {loading ? "AI думает..." : "Суммаризовать AI"}
            </button>

            {/* Модальное окно */}
            {showModal && (
                <div
                    className="modal-overlay"
                    style={{
                        position: "fixed",
                        inset: 0,
                        background: "rgba(0,0,0,0.45)",
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        zIndex: 9999
                    }}
                >
                    <div
                        className="modal-window"
                        style={{
                            background: "#fff",
                            padding: 20,
                            borderRadius: 8,
                            width: "600px",
                            maxHeight: "70vh",
                            overflowY: "auto",
                            position: "relative"
                        }}
                    >
                        <button
                            onClick={() => setShowModal(false)}
                            style={{
                                position: "absolute",
                                right: 10,
                                top: 10,
                                border: "none",
                                background: "transparent",
                                fontSize: 20,
                                cursor: "pointer"
                            }}
                        >
                            ×
                        </button>

                        <h2 style={{ fontSize: "18px", marginBottom: "10px" }}>AI Summary</h2>

                        <pre style={{ whiteSpace: "pre-wrap" }}>
                            {summary}
                        </pre>
                    </div>
                </div>
            )}
        </div>
    );
}
