"""服务状态检测（start_all.bat 的辅助脚本）"""
import socket

ports = {"Go 配送后端": 3000, "FastAPI Agent": 8000, "Vue 前端": 5173}
for name, port in ports.items():
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.settimeout(2)
    try:
        s.connect(("127.0.0.1", port))
        status = "running"
    except:
        status = "offline"
    finally:
        s.close()
    print(f"  {name:<16} (port {port})  {status}")
