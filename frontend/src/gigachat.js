import API from './API.js';

export async function summarizeWithGigachat(markdownText) {
  const res = await API.GIGACHAT_PROXY.post("/api/gigachat/summarize", {
    text: markdownText
  });
  return res.data.summary;
}
