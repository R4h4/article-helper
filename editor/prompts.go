package editor

const (
	EditorPrompt = `As part of a personal exercise, I'm trying to answer the question, "{{QUESTION}}"

1. First, read the following transcription:
<transcription>
%s
</transcription>

2. Clean up the transcription:
   - Remove filler words (um, uh, like, you know, etc.)
   - Correct any obvious grammatical errors
   - Ensure proper capitalization and punctuation
   - Combine fragmented sentences into coherent thoughts
   - Remove any irrelevant or repetitive content

3. Summarize the cleaned-up transcription in bullet points:
   - Identify the main ideas and key points
   - Create concise bullet points that capture the essence of each main idea
   - Ensure the summary is comprehensive yet brief

5. Output your results as a JSON in the following format:
   {
	   "cleaned_transcription": "[Insert the cleaned-up transcription here],
	   "summary": "[Insert bullet point summary here]"
   }
`
)
