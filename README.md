	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ws.EmitToAnyEvent("heartbeat", handlers.BuildBattleIndex(handlers.BattleIndex))
			}
		}

	}()
 