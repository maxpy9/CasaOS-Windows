# Launcher e installer

Questa cartella contiene ciò che è stato scritto da zero per Windows. La documentazione generale è nel [README principale](../README.md).

| Percorso | Contenuto |
|---|---|
| `launcher/` | Finestra con il pulsante Avvia/Ferma, in Go con le API native di Windows. Avvia i sei servizi e li controlla. |
| `installer/casaos.iss` | Script Inno Setup dell'installer. |
| `setup-tools.ps1` | Scarica Go, Inno Setup e la dashboard pubblicata in `..\_tools`. |
| `generate.ps1` | Genera il codice API dei servizi dai file OpenAPI. |
| `build.ps1` | Compila tutto e produce `dist\CasaOS-Setup-<versione>.exe`. |

Opzioni di `CasaOS.exe`:

| Opzione | Effetto |
|---|---|
| `--autostart` | Avvia il server appena si apre la finestra. |
| `--hidden` | Si apre nell'area di notifica, senza finestra. |
| `--quit` | Ferma il server e chiude il launcher già aperto. La usa la disinstallazione. |
| `--prepare` | Esegue una volta ogni servizio, così l'antivirus li analizza subito. La usa l'installazione. |
| `--screenshot <file>` | Salva un'immagine della finestra, per i test. |
