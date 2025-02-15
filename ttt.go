package main

func CheckWinner(game Game, isX bool) int8 {

	var player uint8
	if isX {
		player = 1
	} else {
		player = 2
	}

	lines := [][3]uint8{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
		{0, 4, 8},
		{2, 4, 6},
	}

	for _, line := range lines {
		if checkLine(game, line, player) {
			return int8(player)
		}
	}

	if game.TurnCount >= 10 {
		return 0
	}
	return -1
}

func checkLine(game Game, nums [3]uint8, player uint8) bool {
	return game.Blocks[nums[0]] == player && game.Blocks[nums[1]] == player && game.Blocks[nums[2]] == player
}
