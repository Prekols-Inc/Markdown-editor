import { useState } from "react";
import { BrainCircuit } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
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
        } catch (e) {
            console.error("AI error:", e);
            setSummary("Ошибка при запросе AI.");
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
                    animation: loading ? "pulse 1.2s infinite" : "none",
                }}
                disabled={loading || !current}
                onClick={handleSummarize}
            >
                <BrainCircuit size={20} strokeWidth={1.75} />
                {loading ? "AI думает..." : "AI суммаризация"}
            </button>

            {summary && !loading && (
                <button
                    className="btn"
                    style={{ marginTop: 10 }}
                    onClick={() => setShowModal(true)}
                >
                    Показать результат
                </button>
            )}

            {showModal && (
                <div
                    style={{
                        position: "fixed",
                        inset: 0,
                        background: "rgba(0,0,0,0.45)",
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        zIndex: 9999,
                    }}
                >
                    <div
                        style={{
                            background: "#fff",
                            padding: 24,
                            borderRadius: 12,
                            width: 800,
                            maxHeight: "80vh",
                            overflowY: "auto",
                            position: "relative",
                            boxShadow: "0 16px 40px rgba(0,0,0,0.15)",
                            fontFamily: "system-ui, sans-serif",
                        }}
                    >
                        <button
                            onClick={() => setShowModal(false)}
                            style={{
                                position: "absolute",
                                right: 12,
                                top: 12,
                                border: "none",
                                background: "transparent",
                                fontSize: 20,
                                cursor: "pointer",
                                color: "#555",
                            }}
                        >
                            ×
                        </button>

                        <h2 style={{ fontSize: 20, marginBottom: 16 }}>
                            AI Summary
                        </h2>

                        <div
                            style={{
                                lineHeight: 1.6,
                                fontSize: 16,
                                color: "#333",
                            }}
                        >
                            <ReactMarkdown remarkPlugins={[remarkGfm]}>
                                {summary}
                            </ReactMarkdown>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
