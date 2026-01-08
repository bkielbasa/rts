package campaign

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

type CampaignDef struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Missions    []string `yaml:"missions"`
}

type CampaignProgress struct {
	CompletedMissions  map[string]bool `yaml:"completed_missions"`
	CurrentCampaign    string          `yaml:"current_campaign,omitempty"`
	CurrentMissionIdx  int             `yaml:"current_mission_idx,omitempty"`
	UnlockedBuildings  []string        `yaml:"unlocked_buildings,omitempty"`
	UnlockedUnits      []string        `yaml:"unlocked_units,omitempty"`
}

type Manager struct {
	campaigns    map[string]*CampaignDef
	campaignList []*CampaignDef
	missions     map[string]*Mission
	progress     *CampaignProgress
	progressPath string
	campaignsDir string
}

func NewManager(campaignsDir string) (*Manager, error) {
	m := &Manager{
		campaigns:    make(map[string]*CampaignDef),
		missions:     make(map[string]*Mission),
		campaignsDir: campaignsDir,
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	m.progressPath = filepath.Join(homeDir, ".tanks", "progress.yaml")

	if err := m.loadCampaigns(); err != nil {
		return nil, err
	}

	if err := m.loadProgress(); err != nil {
		m.progress = &CampaignProgress{
			CompletedMissions: make(map[string]bool),
		}
	}

	return m, nil
}

func (m *Manager) loadCampaigns() error {
	entries, err := os.ReadDir(m.campaignsDir)
	if err != nil {
		return fmt.Errorf("failed to read campaigns directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		campaignPath := filepath.Join(m.campaignsDir, entry.Name(), "campaign.yaml")
		if _, err := os.Stat(campaignPath); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(campaignPath)
		if err != nil {
			continue
		}

		var campaign CampaignDef
		if err := yaml.Unmarshal(data, &campaign); err != nil {
			continue
		}

		m.campaigns[campaign.ID] = &campaign
		m.campaignList = append(m.campaignList, &campaign)

		for _, missionFile := range campaign.Missions {
			missionPath := filepath.Join(m.campaignsDir, entry.Name(), missionFile)
			mission, err := LoadMission(missionPath)
			if err != nil {
				continue
			}
			m.missions[mission.ID] = mission
		}
	}

	sort.Slice(m.campaignList, func(i, j int) bool {
		return m.campaignList[i].ID < m.campaignList[j].ID
	})

	return nil
}

func (m *Manager) loadProgress() error {
	data, err := os.ReadFile(m.progressPath)
	if err != nil {
		return err
	}

	var progress CampaignProgress
	if err := yaml.Unmarshal(data, &progress); err != nil {
		return err
	}

	if progress.CompletedMissions == nil {
		progress.CompletedMissions = make(map[string]bool)
	}

	m.progress = &progress
	return nil
}

func (m *Manager) SaveProgress() error {
	dir := filepath.Dir(m.progressPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create progress directory: %w", err)
	}

	data, err := yaml.Marshal(m.progress)
	if err != nil {
		return fmt.Errorf("failed to serialize progress: %w", err)
	}

	if err := os.WriteFile(m.progressPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write progress: %w", err)
	}

	return nil
}

func (m *Manager) GetCampaigns() []*CampaignDef {
	return m.campaignList
}

func (m *Manager) GetCampaign(id string) *CampaignDef {
	return m.campaigns[id]
}

func (m *Manager) GetMission(id string) *Mission {
	return m.missions[id]
}

func (m *Manager) GetCampaignMissions(campaignID string) []*Mission {
	campaign := m.campaigns[campaignID]
	if campaign == nil {
		return nil
	}

	var missions []*Mission
	for _, missionFile := range campaign.Missions {
		missionPath := filepath.Join(m.campaignsDir, campaignID, missionFile)
		mission, err := LoadMission(missionPath)
		if err == nil {
			missions = append(missions, mission)
		}
	}
	return missions
}

func (m *Manager) IsMissionCompleted(missionID string) bool {
	return m.progress.CompletedMissions[missionID]
}

func (m *Manager) IsMissionUnlocked(campaignID, missionID string) bool {
	campaign := m.campaigns[campaignID]
	if campaign == nil {
		return false
	}

	for i, missionFile := range campaign.Missions {
		missionPath := filepath.Join(m.campaignsDir, campaignID, missionFile)
		mission, err := LoadMission(missionPath)
		if err != nil {
			continue
		}

		if mission.ID == missionID {
			if i == 0 {
				return true
			}
			prevMissionPath := filepath.Join(m.campaignsDir, campaignID, campaign.Missions[i-1])
			prevMission, err := LoadMission(prevMissionPath)
			if err != nil {
				return false
			}
			return m.progress.CompletedMissions[prevMission.ID]
		}
	}

	return false
}

func (m *Manager) CompleteMission(missionID string) {
	m.progress.CompletedMissions[missionID] = true
	m.SaveProgress()
}

func (m *Manager) GetNextMission(campaignID, currentMissionID string) *Mission {
	campaign := m.campaigns[campaignID]
	if campaign == nil {
		return nil
	}

	foundCurrent := false
	for _, missionFile := range campaign.Missions {
		missionPath := filepath.Join(m.campaignsDir, campaignID, missionFile)
		mission, err := LoadMission(missionPath)
		if err != nil {
			continue
		}

		if foundCurrent {
			return mission
		}

		if mission.ID == currentMissionID {
			foundCurrent = true
		}
	}

	return nil
}

func (m *Manager) GetFirstMission(campaignID string) *Mission {
	campaign := m.campaigns[campaignID]
	if campaign == nil || len(campaign.Missions) == 0 {
		return nil
	}

	missionPath := filepath.Join(m.campaignsDir, campaignID, campaign.Missions[0])
	mission, err := LoadMission(missionPath)
	if err != nil {
		return nil
	}
	return mission
}

func (m *Manager) ResetProgress() {
	m.progress = &CampaignProgress{
		CompletedMissions: make(map[string]bool),
	}
	m.SaveProgress()
}

func (m *Manager) GetCompletedCount(campaignID string) (completed, total int) {
	campaign := m.campaigns[campaignID]
	if campaign == nil {
		return 0, 0
	}

	total = len(campaign.Missions)
	for _, missionFile := range campaign.Missions {
		missionPath := filepath.Join(m.campaignsDir, campaignID, missionFile)
		mission, err := LoadMission(missionPath)
		if err != nil {
			continue
		}
		if m.progress.CompletedMissions[mission.ID] {
			completed++
		}
	}
	return completed, total
}
