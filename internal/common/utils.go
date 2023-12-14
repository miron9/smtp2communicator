package common

import "strings"

// Splitter splitts a message into defined chunks
//
// This function will take a message and split it into "chunk" sized parts +
// leading and trailing "..." dots (6 characters in total).
//
// Parameters:
//
// - chunk (int): chunk size
// - message (string): a message to split
//
// Returns:
//
// - messages ([]string): slice of chunks
func Splitter(chunk int, message string) (messages []string) {
	// If chunk is 0 or negative or messages shorter than chunk, return the whole message.
	if chunk <= 0 || len(message) < chunk {
		return []string{message}
	}

	var leadingStr, trailingStr string
	start := 0
	end := chunk
	for {
		if end > len(message) {
			end = len(message)
		} else {
			// this is not the end of a message yet so look for place to split
			originalEnd := end
			for {
				// we want to split either on space or newline character if possible
				if message[end] != ' ' && message[end] != '\n' {
					end--
					// if we went back 250 characters and clearly still no good point to split
					// then just split at the original value of "end"
					if originalEnd-end > 250 {
						end = originalEnd
						break
					}

					// if we splitted on new line then the "..." should be
					// also added on line of its own
					if message[end] == '\n' {
						leadingStr = "...\n"
						trailingStr = "\n..."
					} else {
						leadingStr = "..."
						trailingStr = "..."
					}
				} else {
					break
				}
			}
		}

		// if start == end then we reached end of the message
		if start == end {
			break
		}

		subMsg := strings.TrimSpace(message[start:end])

		messages = append(messages, leadingStr+subMsg+trailingStr)

		start = end
		end += chunk
	}

	// trim leading "..." in first message
	messages[0] = messages[0][3:]

	// trim trailing "..." in last message
	lastMessageId := len(messages) - 1
	lastMessageLen := len(messages[lastMessageId])
	messages[lastMessageId] = messages[lastMessageId][0 : lastMessageLen-3]

	return messages
}
