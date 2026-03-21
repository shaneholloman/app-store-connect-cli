-- Usage:
--   osascript /absolute/path/to/get-apple-2fa-code.scpt
--   osascript /absolute/path/to/get-apple-2fa-code.scpt 90
--
-- This script polls the macOS FollowUpUI accessibility tree for a 6-digit
-- Apple two-factor code and prints the first match to stdout.
--
-- Requirements:
--   - The Apple 2FA dialog must be visible on this Mac.
--   - The caller (Terminal/Codex/etc.) must have Accessibility access.

on run argv
	set timeoutSeconds to 60
	if (count of argv) > 0 then
		try
			set timeoutSeconds to (item 1 of argv) as integer
		on error
			error "invalid timeout seconds: " & (item 1 of argv)
		end try
	end if

	set deadlineAt to (current date) + timeoutSeconds
	repeat while (current date) is less than deadlineAt
		set code to my findTwoFactorCode()
		if code is not "" then
			return code
		end if
		delay 1
	end repeat

	error "timed out waiting for a 2FA code in FollowUpUI. Make sure the Apple dialog is visible and Accessibility access is enabled for your terminal."
end run

on findTwoFactorCode()
	try
		tell application "System Events"
			if not (exists process "FollowUpUI") then
				return ""
			end if

			tell process "FollowUpUI"
				repeat with currentWindow in windows
					set code to my scanElement(currentWindow)
					if code is not "" then
						return code
					end if

					try
						set windowElements to entire contents of currentWindow
						repeat with currentElement in windowElements
							set code to my scanElement(currentElement)
							if code is not "" then
								return code
							end if
						end repeat
					end try
				end repeat
			end tell
		end tell
	on error errMsg number errNum
		error "unable to inspect FollowUpUI via Accessibility: " & errMsg number errNum
	end try

	return ""
end findTwoFactorCode

on scanElement(theElement)
	set candidates to {}

	try
		set end of candidates to my normalizeText(value of theElement)
	end try

	try
		set end of candidates to my normalizeText(name of theElement)
	end try

	try
		set end of candidates to my normalizeText(title of theElement)
	end try

	try
		set end of candidates to my normalizeText(description of theElement)
	end try

	repeat with candidateText in candidates
		set code to my extractFirstCode(candidateText as text)
		if code is not "" then
			return code
		end if
	end repeat

	return ""
end scanElement

on normalizeText(candidateValue)
	if candidateValue is missing value then
		return ""
	end if

	if class of candidateValue is list then
		set previousDelimiters to AppleScript's text item delimiters
		set AppleScript's text item delimiters to " "
		set joinedText to candidateValue as text
		set AppleScript's text item delimiters to previousDelimiters
		return joinedText
	end if

	return candidateValue as text
end normalizeText

on extractFirstCode(sourceText)
	set sourceText to sourceText as text
	try
		return do shell script "/bin/echo " & quoted form of sourceText & " | /usr/bin/grep -Eo '[0-9]{6}' | /usr/bin/head -n1"
	on error
		return ""
	end try
end extractFirstCode
