package editor

const (
	EditorPrompt = `You will be given a transcription of someone's speech. Your task is to clean it up, summarize it, and output the result as a JSON.
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

5. Output your results as a JSON with two keys "cleaned_transcription" (string), and "summary" (string). Example:
   {
	   "cleaned_transcription": "Insert the cleaned-up transcription here",
	   "summary": "Insert bullet point summary here"
   }
`
	HeadlinePrompt = `Based on the following summary, create a short, catchy headline with maximum 5 words that could be used as a directory name.
	Please provide the headline as a JSON with a single key "headline" (string). Example:
	{
		"headline": "Insert your headline here"
	}

	<summary>
	%s
	</summary>
	`
)
