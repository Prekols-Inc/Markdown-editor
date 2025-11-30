import axios from "axios";


export async function summarizeWithGigachat(markdownText) {
  const res = await axios.post("http://localhost:8081/api/gigachat/summarize", {
    text: markdownText
  });
  return res.data.summary;
}
