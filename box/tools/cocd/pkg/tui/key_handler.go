package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// DefaultKeyHandler implements the KeyHandler interface
type DefaultKeyHandler struct {
	commands CommandHandlerInterface
}

// NewKeyHandler creates a new key handler
func NewKeyHandler(commands CommandHandlerInterface) KeyHandler {
	return &DefaultKeyHandler{
		commands: commands,
	}
}

// HandleKeyPress handles keyboard input
func (kh *DefaultKeyHandler) HandleKeyPress(msg tea.KeyMsg, app *BubbleApp) (tea.Model, tea.Cmd) {
	// Handle approval confirmation popup keys first
	if app.viewManager.IsShowingApprovalConfirm() {
		return kh.handleApprovalConfirmKeys(msg, app)
	}
	
	// Handle cancel confirmation popup keys next
	if app.viewManager.IsShowingCancelConfirm() {
		return kh.handleCancelConfirmKeys(msg, app)
	}
	
	// If help is showing, any key closes it (except quit keys)
	if app.showHelp {
		return kh.handleHelpKeys(msg, app)
	}
	
	// Handle main view keys
	return kh.handleMainViewKeys(msg, app)
}

func (kh *DefaultKeyHandler) handleApprovalConfirmKeys(msg tea.KeyMsg, app *BubbleApp) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left":
		app.viewManager.SetApprovalSelection(0)
		return app, nil
	case "right":
		app.viewManager.SetApprovalSelection(1)
		return app, nil
	case "enter":
		if app.viewManager.IsApprovalConfirmed() {
			return app, kh.commands.ApproveDeployment(app.ctx, app.viewManager)
		}
		app.viewManager.HideApprovalConfirm()
		return app, nil
	case "esc":
		app.viewManager.HideApprovalConfirm()
		return app, nil
	case "y", "Y":
		return app, kh.commands.ApproveDeployment(app.ctx, app.viewManager)
	case "n", "N":
		app.viewManager.HideApprovalConfirm()
		return app, nil
	default:
		return app, nil
	}
}

func (kh *DefaultKeyHandler) handleCancelConfirmKeys(msg tea.KeyMsg, app *BubbleApp) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left":
		app.viewManager.SetCancelSelection(0)
		return app, nil
	case "right":
		app.viewManager.SetCancelSelection(1)
		return app, nil
	case "enter":
		if app.viewManager.IsCancelConfirmed() {
			return app, kh.commands.CancelWorkflow(app.ctx, app.viewManager)
		}
		app.viewManager.HideCancelConfirm()
		return app, nil
	case "esc":
		app.viewManager.HideCancelConfirm()
		return app, nil
	case "y", "Y":
		return app, kh.commands.CancelWorkflow(app.ctx, app.viewManager)
	case "n", "N":
		app.viewManager.HideCancelConfirm()
		return app, nil
	default:
		return app, nil
	}
}

func (kh *DefaultKeyHandler) handleHelpKeys(msg tea.KeyMsg, app *BubbleApp) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		app.cancel()
		return app, tea.Quit
	default:
		app.showHelp = false
		return app, nil
	}
}

func (kh *DefaultKeyHandler) handleMainViewKeys(msg tea.KeyMsg, app *BubbleApp) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		app.cancel()
		return app, tea.Quit
		
	case "h", "?":
		app.showHelp = !app.showHelp
		return app, nil
		
	case "esc":
		return app, nil
		
	case "t":
		return app.toggleView()
		
	case "r":
		return app.refreshCurrentView()
		
	case "a":
		if app.viewManager.GetCurrentView() == ViewPending {
			return app.showApprovalConfirmation()
		}
		return app, nil
		
	case "c":
		return app.showCancelConfirmation()
		
	case "up", "k":
		return app.moveCursorUp()
		
	case "down", "j":
		return app.moveCursorDown()
		
	case "enter":
		return app, nil
		
	case "o":
		return app, kh.commands.JumpToActions(app.viewManager, app.jobs, app.recentJobs)
		
	case "left":
		return app.navigatePageLeft()
		
	case "right":
		return app.navigatePageRight()
		
	default:
		return app, nil
	}
}