package save

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	MaxSaveSlots = 10
	SaveDirName  = ".tanks/saves"
)

type Manager struct {
	savePath string
}

func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	savePath := filepath.Join(homeDir, SaveDirName)
	if err := os.MkdirAll(savePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create save directory: %w", err)
	}

	return &Manager{savePath: savePath}, nil
}

func (m *Manager) GetSavePath() string {
	return m.savePath
}

func (m *Manager) slotToFilename(slot int) string {
	return filepath.Join(m.savePath, fmt.Sprintf("save_%02d.yaml", slot))
}

func (m *Manager) SaveGame(state *GameState, slot int, name string, missionID string) error {
	if slot < 0 || slot >= MaxSaveSlots {
		return fmt.Errorf("invalid save slot: %d (must be 0-%d)", slot, MaxSaveSlots-1)
	}

	saveFile := &SaveFile{
		Version: SaveVersion,
		Metadata: SaveMetadata{
			Name:      name,
			Timestamp: time.Now(),
			MissionID: missionID,
			PlayTime:  state.GameTime,
		},
		GameState: *state,
	}

	data, err := yaml.Marshal(saveFile)
	if err != nil {
		return fmt.Errorf("failed to serialize save file: %w", err)
	}

	filename := m.slotToFilename(slot)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}

	return nil
}

func (m *Manager) LoadGame(slot int) (*SaveFile, error) {
	if slot < 0 || slot >= MaxSaveSlots {
		return nil, fmt.Errorf("invalid save slot: %d (must be 0-%d)", slot, MaxSaveSlots-1)
	}

	filename := m.slotToFilename(slot)
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("save slot %d is empty", slot)
		}
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}

	var saveFile SaveFile
	if err := yaml.Unmarshal(data, &saveFile); err != nil {
		return nil, fmt.Errorf("failed to parse save file: %w", err)
	}

	if saveFile.Version > SaveVersion {
		return nil, fmt.Errorf("save file version %d is newer than supported version %d", saveFile.Version, SaveVersion)
	}

	return &saveFile, nil
}

func (m *Manager) DeleteSave(slot int) error {
	if slot < 0 || slot >= MaxSaveSlots {
		return fmt.Errorf("invalid save slot: %d (must be 0-%d)", slot, MaxSaveSlots-1)
	}

	filename := m.slotToFilename(slot)
	err := os.Remove(filename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete save file: %w", err)
	}

	return nil
}

type SaveSlotInfo struct {
	Slot     int
	Empty    bool
	Metadata *SaveMetadata
}

func (m *Manager) ListSaves() []SaveSlotInfo {
	slots := make([]SaveSlotInfo, MaxSaveSlots)

	for i := 0; i < MaxSaveSlots; i++ {
		slots[i] = SaveSlotInfo{
			Slot:  i,
			Empty: true,
		}

		filename := m.slotToFilename(i)
		data, err := os.ReadFile(filename)
		if err != nil {
			continue
		}

		var saveFile SaveFile
		if err := yaml.Unmarshal(data, &saveFile); err != nil {
			continue
		}

		slots[i].Empty = false
		slots[i].Metadata = &saveFile.Metadata
	}

	return slots
}

func (m *Manager) GetLatestSave() *SaveSlotInfo {
	slots := m.ListSaves()

	var latest *SaveSlotInfo
	var latestTime time.Time

	for i := range slots {
		if slots[i].Empty {
			continue
		}
		if slots[i].Metadata.Timestamp.After(latestTime) {
			latestTime = slots[i].Metadata.Timestamp
			latest = &slots[i]
		}
	}

	return latest
}

func (m *Manager) GetSavesSortedByTime() []SaveSlotInfo {
	slots := m.ListSaves()

	var nonEmpty []SaveSlotInfo
	for _, slot := range slots {
		if !slot.Empty {
			nonEmpty = append(nonEmpty, slot)
		}
	}

	sort.Slice(nonEmpty, func(i, j int) bool {
		return nonEmpty[i].Metadata.Timestamp.After(nonEmpty[j].Metadata.Timestamp)
	})

	return nonEmpty
}

func (m *Manager) FindEmptySlot() int {
	for i := 0; i < MaxSaveSlots; i++ {
		filename := m.slotToFilename(i)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return i
		}
	}
	return -1
}

func (m *Manager) QuickSave(state *GameState, missionID string) (int, error) {
	slot := m.FindEmptySlot()
	if slot == -1 {
		slots := m.GetSavesSortedByTime()
		if len(slots) > 0 {
			slot = slots[len(slots)-1].Slot
		} else {
			slot = 0
		}
	}

	name := fmt.Sprintf("Quick Save - %s", time.Now().Format("Jan 2 15:04"))
	return slot, m.SaveGame(state, slot, name, missionID)
}

func (m *Manager) QuickLoad() (*SaveFile, error) {
	latest := m.GetLatestSave()
	if latest == nil {
		return nil, fmt.Errorf("no save files found")
	}
	return m.LoadGame(latest.Slot)
}
