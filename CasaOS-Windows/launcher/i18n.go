package main

import "golang.org/x/sys/windows"

var (
	procGetUserDefaultUILanguage = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetUserDefaultUILanguage")

	italian = func() bool {
		lang, _, _ := procGetUserDefaultUILanguage.Call()
		return lang&0x3ff == 0x10 // LANG_ITALIAN
	}()
)

var texts = map[string][2]string{
	"subtitle":        {"Il tuo server personale, solo quando vuoi tu", "Your personal server, only when you want it"},
	"stateStopped":    {"Server spento", "Server stopped"},
	"stateStarting":   {"Avvio in corso…", "Starting…"},
	"stateRunning":    {"Server attivo", "Server running"},
	"stateStopping":   {"Arresto in corso…", "Stopping…"},
	"stateError":      {"Avvio non riuscito", "Could not start"},
	"btnStart":        {"Avvia server", "Start server"},
	"btnStop":         {"Ferma server", "Stop server"},
	"btnStarting":     {"Avvio…", "Starting…"},
	"btnStopping":     {"Arresto…", "Stopping…"},
	"hintStopped":     {"Il server è spento e non usa risorse. Non parte da solo all'avvio di Windows.", "The server is off and uses no resources. It does not start with Windows."},
	"hintRunning":     {"Puoi chiudere questa finestra: il server resta attivo finché non lo fermi.", "You can close this window: the server keeps running until you stop it."},
	"labelLocal":      {"Da questo PC", "On this PC"},
	"labelLAN":        {"Da altri dispositivi in rete", "From other devices on the network"},
	"noAddress":       {"L'indirizzo comparirà qui dopo l'avvio", "The address will appear here once started"},
	"noLAN":           {"Rete non disponibile", "Network not available"},
	"btnOpen":         {"Apri dashboard", "Open dashboard"},
	"btnCopy":         {"Copia link", "Copy link"},
	"btnCopied":       {"Copiato", "Copied"},
	"dockerRunning":   {"Docker Desktop attivo: le app sono disponibili", "Docker Desktop is running: apps are available"},
	"dockerStopped":   {"Docker Desktop spento: le app non sono disponibili", "Docker Desktop is off: apps are not available"},
	"dockerStarting":  {"Docker Desktop in avvio…", "Docker Desktop is starting…"},
	"dockerMissing":   {"Docker Desktop non installato: serve per le app", "Docker Desktop not installed: apps need it"},
	"autoDocker":      {"Avvia Docker Desktop con il server", "Start Docker Desktop with the server"},
	"linkData":        {"Cartella dati", "Data folder"},
	"linkLogs":        {"Log", "Logs"},
	"linkDocker":      {"Scarica Docker", "Get Docker"},
	"menuShow":        {"Apri CasaOS", "Open CasaOS"},
	"menuOpen":        {"Apri dashboard", "Open dashboard"},
	"menuStart":       {"Avvia server", "Start server"},
	"menuStop":        {"Ferma server", "Stop server"},
	"menuQuit":        {"Esci", "Quit"},
	"trayTitle":       {"CasaOS è ancora attivo", "CasaOS is still running"},
	"trayText":        {"Il server continua a funzionare. Puoi fermarlo o riaprire la finestra da qui.", "The server keeps running. You can stop it or reopen the window from here."},
	"starting":        {"Preparazione…", "Getting ready…"},
	"startingService": {"Avvio dei servizi (%d di %d)…", "Starting the services (%d of %d)…"},
	"startingWeb":     {"Avvio della dashboard…", "Starting the dashboard…"},
	"stopping":        {"Arresto dei servizi…", "Stopping the services…"},
	"heldByAntivirus": {"L'antivirus sta analizzando CasaOS, che non è firmato digitalmente. La prima volta può volerci del tempo.", "Your antivirus is analyzing CasaOS, which is not digitally signed. The first time this may take a while."},
	"errFolders":      {"Impossibile creare le cartelle di CasaOS", "Could not create the folders of CasaOS"},
	"errService":      {"Problema con il servizio", "Problem with the service"},
	"errEnded":        {"si è chiuso subito, controlla i log", "it ended right away, please check the logs"},
	"errTimeout":      {"non ha risposto in tempo", "it did not answer in time"},
	"errPort":         {"Impossibile sapere la porta del server", "Could not find out the port of the server"},
	"errWeb":          {"La dashboard non risponde", "The dashboard does not answer"},
	"errCrash":        {"Il servizio %s continua a chiudersi: server fermato. Controlla i log.", "The service %s keeps on ending: server stopped. Please check the logs."},
}

// tr returns a text in the language of Windows: Italian, or English otherwise.
func tr(key string) string {
	t, ok := texts[key]
	if !ok {
		return key
	}
	if italian {
		return t[0]
	}
	return t[1]
}
