$procId = (netstat -ano | Select-String ":8000 " | Select-String "LISTENING").ToString() -split '\s+' | Select-Object -Last 1
Write-Output "Found PID: $procId"
if ($procId -match '^\d+$') {
    Stop-Process -Id $procId -Force -ErrorAction SilentlyContinue
    Write-Output "Killed PID: $procId"
} else {
    Write-Output "No process found on port 8000"
}
