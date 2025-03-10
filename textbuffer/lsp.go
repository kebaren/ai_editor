package textbuffer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// LSPServer represents a connection to a Language Server Protocol server
type LSPServer struct {
	// Command to start the LSP server
	cmd *exec.Cmd
	// Standard input pipe to the LSP server
	stdin io.WriteCloser
	// Standard output pipe from the LSP server
	stdout io.ReadCloser
	// Standard error pipe from the LSP server
	stderr io.ReadCloser
	// Context for cancellation
	ctx context.Context
	// Cancel function for the context
	cancel context.CancelFunc
	// Request ID counter
	nextID int
	// Mutex for thread safety
	mutex sync.Mutex
	// Map of request ID to response channel
	responses map[int]chan json.RawMessage
	// Map of notification method to handler
	notificationHandlers map[string]NotificationHandler
	// TextBuffer reference
	textBuffer *TextBuffer
	// Server capabilities
	capabilities ServerCapabilities
	// Server initialized
	initialized bool
	// Server name
	name string
	// Server root URI
	rootURI string
}

// ServerCapabilities represents the capabilities of an LSP server
type ServerCapabilities struct {
	TextDocumentSync                interface{} `json:"textDocumentSync,omitempty"`
	CompletionProvider              interface{} `json:"completionProvider,omitempty"`
	HoverProvider                   interface{} `json:"hoverProvider,omitempty"`
	SignatureHelpProvider           interface{} `json:"signatureHelpProvider,omitempty"`
	DefinitionProvider              interface{} `json:"definitionProvider,omitempty"`
	ReferencesProvider              interface{} `json:"referencesProvider,omitempty"`
	DocumentSymbolProvider          interface{} `json:"documentSymbolProvider,omitempty"`
	WorkspaceSymbolProvider         interface{} `json:"workspaceSymbolProvider,omitempty"`
	CodeActionProvider              interface{} `json:"codeActionProvider,omitempty"`
	CodeLensProvider                interface{} `json:"codeLensProvider,omitempty"`
	DocumentFormattingProvider      interface{} `json:"documentFormattingProvider,omitempty"`
	DocumentRangeFormattingProvider interface{} `json:"documentRangeFormattingProvider,omitempty"`
	RenameProvider                  interface{} `json:"renameProvider,omitempty"`
}

// NotificationHandler is a function that handles LSP notifications
type NotificationHandler func(method string, params json.RawMessage)

// LSPManager manages LSP servers for the text buffer
type LSPManager struct {
	// Map of language ID to LSP server
	servers map[string]*LSPServer
	// TextBuffer reference
	textBuffer *TextBuffer
	// Mutex for thread safety
	mutex sync.RWMutex
}

// NewLSPManager creates a new LSP manager
func NewLSPManager(textBuffer *TextBuffer) *LSPManager {
	return &LSPManager{
		servers:    make(map[string]*LSPServer),
		textBuffer: textBuffer,
		mutex:      sync.RWMutex{},
	}
}

// StartServer starts an LSP server for the given language ID
func (lm *LSPManager) StartServer(languageID, command string, args ...string) (*LSPServer, error) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// Check if server already exists
	if _, exists := lm.servers[languageID]; exists {
		return nil, fmt.Errorf("LSP server for language '%s' already started", languageID)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create command
	cmd := exec.CommandContext(ctx, command, args...)

	// Create pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Create server
	server := &LSPServer{
		cmd:                  cmd,
		stdin:                stdin,
		stdout:               stdout,
		stderr:               stderr,
		ctx:                  ctx,
		cancel:               cancel,
		nextID:               1,
		mutex:                sync.Mutex{},
		responses:            make(map[int]chan json.RawMessage),
		notificationHandlers: make(map[string]NotificationHandler),
		textBuffer:           lm.textBuffer,
		initialized:          false,
		name:                 languageID,
		rootURI:              "file:///",
	}

	// Start command
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start LSP server: %v", err)
	}

	// Start reading responses
	go server.readResponses()

	// Initialize server
	if err := server.initialize(); err != nil {
		server.Stop()
		return nil, fmt.Errorf("failed to initialize LSP server: %v", err)
	}

	// Add server to manager
	lm.servers[languageID] = server

	return server, nil
}

// StopServer stops an LSP server for the given language ID
func (lm *LSPManager) StopServer(languageID string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	server, exists := lm.servers[languageID]
	if !exists {
		return fmt.Errorf("LSP server for language '%s' not found", languageID)
	}

	// Stop server
	if err := server.Stop(); err != nil {
		return err
	}

	// Remove server from manager
	delete(lm.servers, languageID)

	return nil
}

// GetServer gets an LSP server for the given language ID
func (lm *LSPManager) GetServer(languageID string) (*LSPServer, error) {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	server, exists := lm.servers[languageID]
	if !exists {
		return nil, fmt.Errorf("LSP server for language '%s' not found", languageID)
	}

	return server, nil
}

// GetServers gets all LSP servers
func (lm *LSPManager) GetServers() map[string]*LSPServer {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	// Create a copy of the servers map
	servers := make(map[string]*LSPServer)
	for languageID, server := range lm.servers {
		servers[languageID] = server
	}

	return servers
}

// Close 关闭LSP管理器，清理资源
func (lm *LSPManager) Close() {
	// 关闭所有语言服务器连接
	// 清理临时文件
	// 停止所有相关的goroutine
	lm.textBuffer = nil
}

// LSPServer methods

// initialize initializes the LSP server
func (ls *LSPServer) initialize() error {
	// Send initialize request
	params := map[string]interface{}{
		"processId": nil,
		"rootUri":   ls.rootURI,
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"synchronization": map[string]interface{}{
					"didSave":           true,
					"willSave":          true,
					"change":            2, // Incremental
					"willSaveWaitUntil": true,
				},
				"completion": map[string]interface{}{
					"dynamicRegistration": true,
					"completionItem": map[string]interface{}{
						"snippetSupport":          true,
						"commitCharactersSupport": true,
						"documentationFormat":     []string{"markdown", "plaintext"},
						"deprecatedSupport":       true,
						"preselectSupport":        true,
					},
					"contextSupport": true,
				},
				"hover": map[string]interface{}{
					"dynamicRegistration": true,
					"contentFormat":       []string{"markdown", "plaintext"},
				},
				"signatureHelp": map[string]interface{}{
					"dynamicRegistration": true,
					"signatureInformation": map[string]interface{}{
						"documentationFormat": []string{"markdown", "plaintext"},
						"parameterInformation": map[string]interface{}{
							"labelOffsetSupport": true,
						},
					},
				},
				"definition": map[string]interface{}{
					"dynamicRegistration": true,
					"linkSupport":         true,
				},
				"references": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"documentSymbol": map[string]interface{}{
					"dynamicRegistration": true,
					"symbolKind": map[string]interface{}{
						"valueSet": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
					},
					"hierarchicalDocumentSymbolSupport": true,
				},
				"formatting": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"rangeFormatting": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"rename": map[string]interface{}{
					"dynamicRegistration": true,
					"prepareSupport":      true,
				},
				"codeAction": map[string]interface{}{
					"dynamicRegistration": true,
					"codeActionLiteralSupport": map[string]interface{}{
						"codeActionKind": map[string]interface{}{
							"valueSet": []string{
								"",
								"quickfix",
								"refactor",
								"refactor.extract",
								"refactor.inline",
								"refactor.rewrite",
								"source",
								"source.organizeImports",
							},
						},
					},
				},
			},
			"workspace": map[string]interface{}{
				"symbol": map[string]interface{}{
					"dynamicRegistration": true,
					"symbolKind": map[string]interface{}{
						"valueSet": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
					},
				},
			},
		},
	}

	// Send initialize request
	var result struct {
		Capabilities ServerCapabilities `json:"capabilities"`
	}
	if err := ls.sendRequest("initialize", params, &result); err != nil {
		return err
	}

	// Store server capabilities
	ls.capabilities = result.Capabilities

	// Send initialized notification
	if err := ls.sendNotification("initialized", map[string]interface{}{}); err != nil {
		return err
	}

	ls.initialized = true

	return nil
}

// Stop stops the LSP server
func (ls *LSPServer) Stop() error {
	// Send shutdown request
	if ls.initialized {
		if err := ls.sendRequest("shutdown", nil, nil); err != nil {
			return err
		}

		// Send exit notification
		if err := ls.sendNotification("exit", nil); err != nil {
			return err
		}
	}

	// Cancel context
	ls.cancel()

	// Wait for command to exit
	if err := ls.cmd.Wait(); err != nil {
		// Ignore error if context was canceled
		if ls.ctx.Err() != nil {
			return nil
		}
		return err
	}

	return nil
}

// RegisterNotificationHandler registers a handler for LSP notifications
func (ls *LSPServer) RegisterNotificationHandler(method string, handler NotificationHandler) {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	ls.notificationHandlers[method] = handler
}

// UnregisterNotificationHandler unregisters a handler for LSP notifications
func (ls *LSPServer) UnregisterNotificationHandler(method string) {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	delete(ls.notificationHandlers, method)
}

// DidOpen notifies the LSP server that a document has been opened
func (ls *LSPServer) DidOpen(uri, languageID, text string) error {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        uri,
			"languageId": languageID,
			"version":    1,
			"text":       text,
		},
	}
	return ls.sendNotification("textDocument/didOpen", params)
}

// DidChange notifies the LSP server that a document has changed
func (ls *LSPServer) DidChange(uri string, version int, changes []TextDocumentContentChangeEvent) error {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":     uri,
			"version": version,
		},
		"contentChanges": changes,
	}
	return ls.sendNotification("textDocument/didChange", params)
}

// DidClose notifies the LSP server that a document has been closed
func (ls *LSPServer) DidClose(uri string) error {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
	}
	return ls.sendNotification("textDocument/didClose", params)
}

// Completion requests completion items at a given position
func (ls *LSPServer) Completion(uri string, line, character int) (interface{}, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"position": map[string]interface{}{
			"line":      line,
			"character": character,
		},
	}
	var result interface{}
	if err := ls.sendRequest("textDocument/completion", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Hover requests hover information at a given position
func (ls *LSPServer) Hover(uri string, line, character int) (interface{}, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"position": map[string]interface{}{
			"line":      line,
			"character": character,
		},
	}
	var result interface{}
	if err := ls.sendRequest("textDocument/hover", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Definition requests the definition of a symbol at a given position
func (ls *LSPServer) Definition(uri string, line, character int) (interface{}, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"position": map[string]interface{}{
			"line":      line,
			"character": character,
		},
	}
	var result interface{}
	if err := ls.sendRequest("textDocument/definition", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// References requests references to a symbol at a given position
func (ls *LSPServer) References(uri string, line, character int, includeDeclaration bool) (interface{}, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"position": map[string]interface{}{
			"line":      line,
			"character": character,
		},
		"context": map[string]interface{}{
			"includeDeclaration": includeDeclaration,
		},
	}
	var result interface{}
	if err := ls.sendRequest("textDocument/references", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// DocumentSymbol requests document symbols
func (ls *LSPServer) DocumentSymbol(uri string) (interface{}, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
	}
	var result interface{}
	if err := ls.sendRequest("textDocument/documentSymbol", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Formatting requests document formatting
func (ls *LSPServer) Formatting(uri string, tabSize int, insertSpaces bool) (interface{}, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"options": map[string]interface{}{
			"tabSize":      tabSize,
			"insertSpaces": insertSpaces,
		},
	}
	var result interface{}
	if err := ls.sendRequest("textDocument/formatting", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// RangeFormatting requests document range formatting
func (ls *LSPServer) RangeFormatting(uri string, startLine, startCharacter, endLine, endCharacter, tabSize int, insertSpaces bool) (interface{}, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"range": map[string]interface{}{
			"start": map[string]interface{}{
				"line":      startLine,
				"character": startCharacter,
			},
			"end": map[string]interface{}{
				"line":      endLine,
				"character": endCharacter,
			},
		},
		"options": map[string]interface{}{
			"tabSize":      tabSize,
			"insertSpaces": insertSpaces,
		},
	}
	var result interface{}
	if err := ls.sendRequest("textDocument/rangeFormatting", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CodeAction requests code actions
func (ls *LSPServer) CodeAction(uri string, startLine, startCharacter, endLine, endCharacter int, diagnostics []interface{}, only []string) (interface{}, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"range": map[string]interface{}{
			"start": map[string]interface{}{
				"line":      startLine,
				"character": startCharacter,
			},
			"end": map[string]interface{}{
				"line":      endLine,
				"character": endCharacter,
			},
		},
		"context": map[string]interface{}{
			"diagnostics": diagnostics,
		},
	}
	if len(only) > 0 {
		params["context"].(map[string]interface{})["only"] = only
	}
	var result interface{}
	if err := ls.sendRequest("textDocument/codeAction", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Rename requests renaming a symbol
func (ls *LSPServer) Rename(uri string, line, character int, newName string) (interface{}, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"position": map[string]interface{}{
			"line":      line,
			"character": character,
		},
		"newName": newName,
	}
	var result interface{}
	if err := ls.sendRequest("textDocument/rename", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// TextDocumentContentChangeEvent represents a change to a text document
type TextDocumentContentChangeEvent struct {
	// Range is the range of the document that changed
	Range *struct {
		Start struct {
			Line      int `json:"line"`
			Character int `json:"character"`
		} `json:"start"`
		End struct {
			Line      int `json:"line"`
			Character int `json:"character"`
		} `json:"end"`
	} `json:"range,omitempty"`
	// RangeLength is the length of the range that got replaced
	RangeLength int `json:"rangeLength,omitempty"`
	// Text is the new text of the range/document
	Text string `json:"text"`
}

// Internal methods

// sendRequest sends a request to the LSP server and waits for a response
func (ls *LSPServer) sendRequest(method string, params interface{}, result interface{}) error {
	ls.mutex.Lock()
	id := ls.nextID
	ls.nextID++
	responseChan := make(chan json.RawMessage, 1)
	ls.responses[id] = responseChan
	ls.mutex.Unlock()

	// Create request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		request["params"] = params
	}

	// Send request
	if err := ls.sendMessage(request); err != nil {
		ls.mutex.Lock()
		delete(ls.responses, id)
		ls.mutex.Unlock()
		return err
	}

	// Wait for response
	select {
	case <-ls.ctx.Done():
		ls.mutex.Lock()
		delete(ls.responses, id)
		ls.mutex.Unlock()
		return errors.New("context canceled")
	case response := <-responseChan:
		ls.mutex.Lock()
		delete(ls.responses, id)
		ls.mutex.Unlock()

		// Check for error
		var errorResponse struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(response, &errorResponse); err == nil && errorResponse.Error.Message != "" {
			return fmt.Errorf("LSP error: %s (code %d)", errorResponse.Error.Message, errorResponse.Error.Code)
		}

		// Unmarshal result
		if result != nil {
			var resultResponse struct {
				Result json.RawMessage `json:"result"`
			}
			if err := json.Unmarshal(response, &resultResponse); err != nil {
				return fmt.Errorf("failed to unmarshal response: %v", err)
			}
			if err := json.Unmarshal(resultResponse.Result, result); err != nil {
				return fmt.Errorf("failed to unmarshal result: %v", err)
			}
		}

		return nil
	}
}

// sendNotification sends a notification to the LSP server
func (ls *LSPServer) sendNotification(method string, params interface{}) error {
	// Create notification
	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if params != nil {
		notification["params"] = params
	}

	// Send notification
	return ls.sendMessage(notification)
}

// sendMessage sends a message to the LSP server
func (ls *LSPServer) sendMessage(message interface{}) error {
	// Marshal message
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// Create content length header
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))

	// Send header and message
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	if _, err := ls.stdin.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}
	if _, err := ls.stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	return nil
}

// readResponses reads responses from the LSP server
func (ls *LSPServer) readResponses() {
	defer func() {
		// Cancel context when done
		ls.cancel()
	}()

	// Read responses
	for {
		// Read content length header
		var contentLength int
		for {
			// Read header line
			var line string
			for {
				b := make([]byte, 1)
				if _, err := ls.stdout.Read(b); err != nil {
					if err == io.EOF {
						return
					}
					continue
				}
				if b[0] == '\n' {
					break
				}
				if b[0] != '\r' {
					line += string(b)
				}
			}

			// Check if header is empty
			if line == "" {
				break
			}

			// Parse content length
			if len(line) > 16 && line[:16] == "Content-Length: " {
				fmt.Sscanf(line[16:], "%d", &contentLength)
			}
		}

		// Read message
		if contentLength > 0 {
			data := make([]byte, contentLength)
			if _, err := io.ReadFull(ls.stdout, data); err != nil {
				if err == io.EOF {
					return
				}
				continue
			}

			// Parse message
			var message map[string]json.RawMessage
			if err := json.Unmarshal(data, &message); err != nil {
				continue
			}

			// Check if message is a response
			if id, ok := message["id"]; ok {
				// Get response channel
				var idValue int
				if err := json.Unmarshal(id, &idValue); err != nil {
					continue
				}

				ls.mutex.Lock()
				responseChan, ok := ls.responses[idValue]
				ls.mutex.Unlock()
				if ok {
					// Send response
					responseChan <- data
				}
			} else if method, ok := message["method"]; ok {
				// Check if message is a notification
				var methodValue string
				if err := json.Unmarshal(method, &methodValue); err != nil {
					continue
				}

				// Get notification handler
				ls.mutex.Lock()
				handler, ok := ls.notificationHandlers[methodValue]
				ls.mutex.Unlock()
				if ok {
					// Call handler
					var params json.RawMessage
					if params, ok = message["params"]; !ok {
						params = json.RawMessage("{}")
					}
					handler(methodValue, params)
				}
			}
		}
	}
}
