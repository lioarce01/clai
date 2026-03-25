package tui

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/lioarce01/clai/internal/config"
	"github.com/rivo/tview"
)

// buildSettingsModal constructs a tview Form for editing application settings.
// onSave is called with the updated config when the user submits.
// onClose is called when the user cancels/closes.
func buildSettingsModal(
	cfg *config.Config,
	onSave func(cfg *config.Config),
	onClose func(),
) *tview.Frame {
	// Local mutable copies of field values
	apiKey := cfg.API.APIKey
	baseURL := cfg.API.BaseURL
	modelName := cfg.Model.Name
	temperature := fmt.Sprintf("%.2f", cfg.Model.Temperature)
	maxTokens := fmt.Sprintf("%d", cfg.Model.MaxTokens)
	topP := fmt.Sprintf("%.2f", cfg.Model.TopP)
	systemPrompt := cfg.Model.SystemPrompt

	errView := tview.NewTextView()
	errView.SetDynamicColors(true)
	errView.SetBackgroundColor(tcell.ColorDefault)
	errView.SetTextColor(tcell.ColorDefault)
	errView.SetText("")

	form := tview.NewForm()
	form.SetBorder(false)
	form.SetBackgroundColor(tcell.ColorDefault)
	form.SetFieldBackgroundColor(tcell.ColorDefault)
	form.SetFieldTextColor(tcell.ColorDefault)
	form.SetLabelColor(tcell.ColorDefault)
	form.SetButtonBackgroundColor(tcell.ColorDefault)
	form.SetButtonTextColor(tcell.ColorDefault)

	form.AddPasswordField("API Key", apiKey, 40, '*', func(text string) {
		apiKey = text
	})
	form.AddInputField("Base URL", baseURL, 40, nil, func(text string) {
		baseURL = text
	})
	form.AddInputField("Model", modelName, 40, nil, func(text string) {
		modelName = text
	})
	form.AddInputField("Temperature", temperature, 10, nil, func(text string) {
		temperature = text
	})
	form.AddInputField("Max Tokens", maxTokens, 10, nil, func(text string) {
		maxTokens = text
	})
	form.AddInputField("Top P", topP, 10, nil, func(text string) {
		topP = text
	})
	form.AddInputField("System Prompt", systemPrompt, 40, nil, func(text string) {
		systemPrompt = text
	})

	form.AddButton("Save", func() {
		newCfg, err := buildConfig(cfg, apiKey, baseURL, modelName, temperature, maxTokens, topP, systemPrompt)
		if err != nil {
			errView.SetText("[red]Error: " + err.Error() + "[-]")
			return
		}
		errView.SetText("")
		onSave(newCfg)
	})
	form.AddButton("Cancel", func() {
		onClose()
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(errView, 1, 0, false)

	frame := tview.NewFrame(flex).
		SetBorders(1, 1, 1, 1, 2, 2)
	frame.SetTitle(" Settings ")
	frame.SetBorder(true)
	frame.SetBackgroundColor(tcell.ColorDefault)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			onClose()
			return nil
		}
		return event
	})

	return frame
}

func buildConfig(
	base *config.Config,
	apiKey, baseURL, modelName, temperature, maxTokens, topP, systemPrompt string,
) (*config.Config, error) {
	temp, err := strconv.ParseFloat(temperature, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid temperature: %w", err)
	}
	maxTok, err := strconv.Atoi(maxTokens)
	if err != nil {
		return nil, fmt.Errorf("invalid max tokens: %w", err)
	}
	topPVal, err := strconv.ParseFloat(topP, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid top P: %w", err)
	}

	cfg := *base
	cfg.API.APIKey = apiKey
	cfg.API.BaseURL = baseURL
	cfg.Model.Name = modelName
	cfg.Model.Temperature = temp
	cfg.Model.MaxTokens = maxTok
	cfg.Model.TopP = topPVal
	cfg.Model.SystemPrompt = systemPrompt
	return &cfg, nil
}
