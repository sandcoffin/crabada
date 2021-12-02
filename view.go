package main

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (et *etubot) gas(msg *tb.Message) {

	et.gasMu.RLock()
	gasPrice := et.gasPrice
	et.gasMu.RUnlock()

	et.bot.Reply(msg, fmt.Sprintf("%d gwei", big.NewInt(0).Div(gasPrice, big.NewInt(1e9))))
}

func (et *etubot) sendTeams(msg *tb.Message) {

	teams, err := et.allTeams()
	if err != nil {
		et.sendError(fmt.Errorf("error fetching teams: %v", err))
		return
	}

	sb := fmt.Sprintf("%d teams\n---------------------\n", len(teams))

	for _, team := range teams {
		teamSummary := fmt.Sprintf("ID: %d\n", team.ID)
		teamSummary += fmt.Sprintf("Strength: %d\n", team.Strength)
		teamSummary += fmt.Sprintf("Account: %s\n", team.Wallet[:7])
		teamSummary += fmt.Sprintf("Status: %s\n", strings.ToLower(team.Status))

		sb += "\n" + teamSummary
	}

	et.bot.Reply(msg, sb)
}

func (et *etubot) sendActiveLoots(msg *tb.Message) {

	games, err := et.activeLoots()
	if err != nil {
		et.sendError(fmt.Errorf("error fetching active loots: %v", err))
		return
	}
	sb := fmt.Sprintf("%d active loots 💰🤑💰\n-------------------------\n", len(games))
	for _, game := range games {
		lootSummary := "💰 Loot\n"
		lootSummary += fmt.Sprintf("Game: %d\n", game.ID)
		lootSummary += fmt.Sprintf("Team: %d\n", game.AttackTeamID)
		lootSummary += fmt.Sprintf("Account: %s\n", game.AttackTeamOwner[:7])
		if len(game.Process) > 0 {
			latestProcess := game.lastProcess()
			var lastAction, status string
			if latestProcess.Action == actionAttack || latestProcess.Action == actionReinforceAttack {

				if latestProcess.Action == actionReinforceAttack {
					lastAction = "reinforced"
				} else {
					lastAction = "attacked"
				}

				if time.Since(latestProcess.txTime()) > processIntervals {
					status = "won"
				} else if len(game.Process) == 6 { // game completed
					status = "waiting to settle"
				} else {
					status = "waiting for opponent's reinforcement"
				}
			} else if latestProcess.Action == actionReinforceDefense {
				lastAction = "opponent reinforced"
				if time.Since(latestProcess.txTime()) > processIntervals {
					status = "lost"
				} else {
					status = fmt.Sprintf("%d reinforcement needed", game.DefensePoint-game.AttackPoint)
				}
			}
			lootSummary += fmt.Sprintf("Last action: %s\n", lastAction)
			lootSummary += fmt.Sprintf("Status: %s\n", status)
		}
		if game.canSettle() {
			lootSummary += "Ready: yes\n"
		} else {
			lootSummary += fmt.Sprintf("Settle in: %s\n", time.Until(game.settleTime()))
		}

		sb += "\n" + lootSummary
	}

	et.bot.Reply(msg, sb)
}
