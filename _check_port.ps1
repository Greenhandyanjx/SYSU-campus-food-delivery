try {
    $conn = Get-NetTCPConnection -LocalPort 8000 -ErrorAction Stop
    Write-Output "OwningProcess: $($conn.OwningProcess)"
    Write-Output "State: $($conn.State)"
    $proc = Get-Process -Id $conn.OwningProcess -ErrorAction SilentlyContinue
    if ($proc) {
        Write-Output "ProcessName: $($proc.ProcessName)"
        Write-Output "StartTime: $($proc.StartTime)"
        Write-Output "SessionId: $($proc.SessionId)"
    } else {
        Write-Output "Process not found (PID: $($conn.OwningProcess))"
    }
} catch {
    Write-Output "No TCP connection on port 8000"
}
