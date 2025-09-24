# Technical Specification: `overheard`

**Version:** 1.0
**Author:** Gemini

## 1. Overview

`overheard` is a command-line utility written in Go for scheduling and recording audio from internet radio streams.

The application reads a configuration file that specifies a stream URL (`.pls` or `.m3u`), a start time, a recording duration, and an output directory. At the designated time, it connects to the stream, downloads the audio, encodes it to the FLAC format, and saves it to the specified directory.

A primary technical requirement is to use only native Go libraries, avoiding any dependency on C/C++ compilers (no CGO).

## 2. Core Features

-   **Scheduled Recording:** Initiate a recording at a specific time and for a specific duration defined by the user.
-   **Configuration Driven:** All parameters are managed via a simple YAML configuration file.
-   **Playlist Parsing:** Automatically parse `.pls` and `.m3u` playlist files to find the direct audio stream URL.
-   **FLAC Encoding:** Transcode the incoming audio stream (e.g., MP3, AAC) into the lossless FLAC format in real-time.
-   **Pure Go:** Built with a toolchain that does not require CGO, ensuring easy compilation and cross-platform compatibility.
-   **Robust Error Handling:** Provide clear, actionable error messages for issues like invalid configuration, network problems, or parsing failures.

## 3. Command-Line Interface (CLI) Design

The application will be built using the `spf13/cobra` library to provide a clean and extensible command structure.

```
overheard [command]
```

### Commands

-   `overheard record`:
    -   This is the primary command.
    -   It locates and reads the configuration file.
    -   It waits until the scheduled `start_time`.
    -   It executes the recording process as defined in the configuration.
    -   It will exit after the recording is complete or if a critical error occurs.

-   `overheard config`:
    -   Displays the location where the application expects to find the configuration file.
    -   Validates the syntax of the existing configuration file and prints a success message or a detailed error.

-   `overheard version`:
    -   Prints the current version of the application.

-   `overheard add`:
    -   Interactively prompts for required fields and adds a new entry to the config file
    - Should instruct the user at each prompt with an example format

## 4. Configuration File

The application will be configured via a file named `config.yaml`.

### Location

The application will look for the configuration file in a platform-specific user config directory. This will be handled by Go's `os.UserConfigDir()`.

-   **Linux:** `~/.config/overheard/config.yaml`
-   **macOS:** `~/Library/Application Support/overheard/config.yaml`
-   **Windows:** `%APPDATA%\overheard\config.yaml`

### Format and Fields

The configuration will be in YAML format for readability.

```yaml
# The URL of the radio stream playlist.
# Must point to a .pls or .m3u file.
stream_url: "http://radio.example.com/stream.pls"

# The scheduled start time for the recording is in a simple "Thu 8:05pm" format.
start_time: "Wed 6:00pm"

# The duration of the recording.
# Format is a string parsable by Go's time.ParseDuration (e.g., "1h30m", "45m").
duration: "2h"

# The absolute path to the directory where the .flac file will be saved.
# The directory must exist.
output_dir: "/home/user/recordings"
```

## 5. Application Logic & Workflow (`record` command)

1.  **Load Configuration:** The app starts and uses the `spf13/viper` library to find and unmarshal `config.yaml`.
2.  **Validate Configuration:** All fields are validated. If `stream_url` is empty, `start_time` is in the past, `duration` is invalid, or `output_dir` does not exist, the app will exit with a descriptive error.
3.  **Schedule Wait:** The app calculates the duration between the current time and `start_time`. It will then print a message "Waiting until [start_time] to begin recording..." and enter a sleep state (`time.Sleep`).
4.  **Fetch Playlist:** Once the wait is over, the app makes an HTTP GET request to the `stream_url`.
5.  **Parse Playlist:**
    -   The response body is scanned line by line.
    -   The logic will check for patterns matching `.pls` (`File1=...`) or `.m3u` (lines that are not comments starting with `#`).
    -   The first valid URL found will be extracted as the direct audio stream URL. If no URL is found, the app exits with an error.
6.  **Initiate Recording:**
    -   A `context.WithTimeout` is created using the `duration` from the config to ensure the recording stops after the specified time.
    -   An HTTP GET request is made to the extracted audio stream URL. This provides a streaming response body (`io.ReadCloser`).
7.  **Setup Encoding Pipeline:**
    -   An output file is created in the `output_dir`. The filename will be a timestamp of the start time, e.g., `2025-10-15T20-00-00-0500.flac`.
    -   The streaming HTTP body is fed into a **decoder** (e.g., a pure Go MP3 decoder).
    -   The decoder's output (raw PCM audio data) is fed into a pure Go **FLAC encoder**.
    -   The encoder's output is written to the output file.
    -   This pipeline `[HTTP Body -> Decoder -> Encoder -> File]` ensures that the entire stream is not held in memory.
8.  **Graceful Shutdown:**
    -   When the `context` times out (recording duration is met), the HTTP connection is closed, the encoder is finalized, and the file is closed.
    -   The application prints a "Recording complete" message and exits.

## 6. Proposed Go Libraries & Packages (Native Go)

-   **CLI Framework:** `github.com/spf13/cobra` - Industry standard for Go CLIs.
-   **Configuration:** `github.com/spf13/viper` - For loading configuration from file and environment.
-   **Playlist Parsing:** Standard Library (`net/http`, `bufio.Scanner`). The formats are simple enough that a dedicated library is not required.
-   **Audio Decoding:** `github.com/hajimehoshi/go-mp3`. A robust, pure-Go MP3 decoder. Support for other formats like AAC will require sourcing other pure-Go libraries and can be added later.
-   **Audio Encoding:** `github.com/mewkiz/flac`. A native Go FLAC encoder.

## 7. Error Handling

-   **Config Errors:** App will exit with code 1 and print messages like "Configuration file not found at [path]" or "Invalid 'start_time': must be in the future."
-   **Network Errors:** Retry logic will not be implemented in v1.0. If a network connection fails during playlist fetching or streaming, the app will exit with an error message.
-   **Parsing/Encoding Errors:** If the stream is not in a supported format or is corrupted, the app will clean up the partially written file and exit with an error detailing the failure.

## 8. Open Questions

-   **Initial Audio Format Support:** The initial implementation will focus on decoding MP3 streams due to the availability of high-quality pure Go libraries. Support for AAC or other formats can be added in future versions.
-   **Stream Reconnection:** If the stream drops mid-recording, v1.0 will not attempt to reconnect. The recording will terminate.
