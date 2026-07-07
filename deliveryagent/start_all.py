"""JD-Agent 全栈服务启动器 (Python)

用法:
  python start_all.py           交互模式（启动服务 + 菜单）
  python start_all.py --start   仅启动服务（供 bat 调用）
  python start_all.py --status  显示服务状态
  python start_all.py --docker  开关 Docker
  python start_all.py --stop    停止所有服务
"""
import socket
import subprocess
import sys
import time
import os
import signal

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
GO_BACKEND_DIR = r"D:\桌面\软件工程中级实训\SYSU-campus-food-delivery\backend"
GO_BACKEND = os.path.join(GO_BACKEND_DIR, "server.exe")
FRONTEND_DIR = r"D:\桌面\软件工程中级实训\SYSU-campus-food-delivery\frontend"
DOCKER_COMPOSE = os.path.join(BASE_DIR, "docker-compose.yml")          # 完整部署（api+ui+infra）
DOCKER_COMPOSE_INFRA = os.path.join(BASE_DIR, "docker-compose.infra.yml")  # 仅基础设施
DOCKER_MIRROR = os.path.join(BASE_DIR, "docker-compose.mirror.yml")
DOCKER_COMPOSE_CMD = None  # 延迟初始化

PROCESSES = []
DOCKER_RUNNING = False


# ── 端口/健康检查 ──

def port_free(port: int, timeout: float = 2) -> bool:
    s = socket.socket()
    s.settimeout(timeout)
    try:
        s.connect(("127.0.0.1", port))
        return False
    except:
        return True
    finally:
        s.close()


def wait_port(port: int, timeout: float = 10) -> bool:
    deadline = time.time() + timeout
    while time.time() < deadline:
        if not port_free(port):
            return True
        time.sleep(0.5)
    return False


def check_health(url: str, timeout: float = 3) -> bool:
    import urllib.request
    try:
        resp = urllib.request.urlopen(url, timeout=timeout)
        return resp.status == 200
    except:
        return False


# ── 进程管理 ──

def kill_port(port: int):
    """杀掉占用指定端口的进程"""
    try:
        result = subprocess.run(
            f'netstat -ano | findstr ":{port} " | findstr LISTEN',
            shell=True, capture_output=True, text=True, timeout=5
        )
        killed = False
        for line in result.stdout.strip().split("\n"):
            parts = line.strip().split()
            if len(parts) >= 5:
                pid = int(parts[4])
                try:
                    os.kill(pid, signal.SIGTERM)
                    time.sleep(0.3)
                    killed = True
                except:
                    pass
        if killed:
            print(f"  已释放端口 {port}")
    except:
        pass


def cleanup():
    """停止本地服务（进程列表 + 端口扫描）"""
    print("\n正在停止本地服务...")
    for proc in PROCESSES:
        if proc.poll() is None:
            proc.terminate()
    time.sleep(1)
    for proc in PROCESSES:
        if proc.poll() is None:
            proc.kill()
    # 兜底：端口扫描确保释放
    for p in (3000, 8000, 5173):
        kill_port(p)
    print("本地服务已停止。")


# ── Docker ──

def docker_check() -> bool:
    r = subprocess.run("docker --version", shell=True, capture_output=True, text=True, timeout=10)
    if r.returncode == 0:
        print(f"  Docker: {r.stdout.strip()}")
        return True
    print("  [警告] Docker 未安装或未运行")
    return False


def _docker_compose_cmd() -> list:
    """构建 docker compose 命令（含 mirror override），优先使用 infra-only compose"""
    global DOCKER_COMPOSE_CMD
    if DOCKER_COMPOSE_CMD is None:
        # 优先用 infra-only（不启动 api/ui，避免与本地进程端口冲突）
        compose_file = DOCKER_COMPOSE_INFRA if os.path.exists(DOCKER_COMPOSE_INFRA) else DOCKER_COMPOSE
        cmd = ["docker", "compose", "-f", compose_file]
        if os.path.exists(DOCKER_MIRROR):
            cmd.extend(["-f", DOCKER_MIRROR])
        DOCKER_COMPOSE_CMD = cmd
    return DOCKER_COMPOSE_CMD


def _pre_pull_images():
    """通过 DaoCloud 镜像预拉取基础设施镜像"""
    mirror_prefix = "docker.m.daocloud.io/library"
    images = {
        "redis:7-alpine": f"{mirror_prefix}/redis:7-alpine",
        "postgres:16-alpine": f"{mirror_prefix}/postgres:16-alpine",
        "rabbitmq:3-management-alpine": f"{mirror_prefix}/rabbitmq:3-management-alpine",
    }
    for tag, mirror_url in images.items():
        r = subprocess.run(
            ["docker", "image", "inspect", tag],
            capture_output=True, timeout=10
        )
        if r.returncode == 0:
            continue  # 已存在
        print(f"  预拉取 {tag} ...")
        pull = subprocess.run(
            ["docker", "pull", mirror_url],
            capture_output=True, timeout=120
        )
        if pull.returncode == 0:
            subprocess.run(["docker", "tag", mirror_url, tag],
                           capture_output=True, timeout=10)
            print(f"    OK")
        else:
            err = pull.stderr.decode("utf-8", errors="replace")[:200]
            print(f"    FAIL: {err}")


def _docker_run(args: list, timeout: int = 60) -> subprocess.CompletedProcess:
    """执行 docker 命令，避免 GBK 编码崩溃"""
    try:
        return subprocess.run(
            args, shell=False,
            stdout=subprocess.PIPE, stderr=subprocess.PIPE,
            timeout=timeout
        )
    except subprocess.TimeoutExpired:
        print(f"  [超时] Docker 命令执行超过 {timeout}s")
        return subprocess.CompletedProcess(args, -1, b"", b"timeout")


def docker_up() -> bool:
    global DOCKER_RUNNING
    compose_file = DOCKER_COMPOSE_INFRA if os.path.exists(DOCKER_COMPOSE_INFRA) else DOCKER_COMPOSE
    if not os.path.exists(compose_file):
        print(f"  [错误] 未找到 {compose_file}")
        return False
    print(f"  [Docker] 使用 {os.path.basename(compose_file)}")
    print("  [Docker] 预拉取镜像（DaoCloud 加速）...")
    _pre_pull_images()
    print("  [Docker] 启动容器中...")
    r = _docker_run(_docker_compose_cmd() + ["up", "-d"], timeout=120)
    if r.returncode == 0:
        print("  [Docker] 已启动 OK")
        DOCKER_RUNNING = True
        return True
    else:
        err = r.stderr.decode("utf-8", errors="replace")[:500] if r.stderr else ""
        print(f"  [Docker] 启动失败:\n    {err}")
        return False


def docker_stop():
    global DOCKER_RUNNING
    print("  [Docker] 停止容器中...")
    # 直接用容器名 stop（兼容不同 compose 来源启动的容器）
    for c in ("jd-agent-redis", "jd-agent-postgres", "jd-agent-rabbitmq"):
        subprocess.run(
            ["docker", "stop", c],
            capture_output=True, timeout=30
        )
        subprocess.run(
            ["docker", "rm", c],
            capture_output=True, timeout=30
        )
    DOCKER_RUNNING = False
    print("  [Docker] 已停止")


def docker_ps():
    r = _docker_run(
        ["docker", "ps", "--filter", "name=jd-agent",
         "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}"],
        timeout=10
    )
    if r.returncode == 0 and r.stdout:
        out = r.stdout.decode("utf-8", errors="replace").strip()
        for line in out.split("\n"):
            print(f"  {line}")
    else:
        print("  (无运行中的容器)")


def docker_auto_start() -> bool:
    """检查 Docker 基础设施是否运行，未运行则自动启动"""
    if not docker_check():
        print("  [跳过] Docker 不可用，跳过基础设施启动")
        return False
    # 检查 jd-agent 开头的容器是否在运行
    r = subprocess.run(
        ["docker", "ps", "--filter", "name=jd-agent", "-q"],
        capture_output=True, timeout=10
    )
    if r.returncode == 0 and r.stdout.strip():
        print("  [Docker] 基础设施容器已在运行")
        return True
    print("  [Docker] 基础设施未运行，自动启动...")
    return docker_up()


# ── 服务启动 ──

def start_service(cmd: list, name: str, port: int, cwd: str = None,
                  health_url: str = None, wait_seconds: float = 4) -> bool:
    print(f"[{name}] 启动中...")
    kill_port(port)
    proc = subprocess.Popen(
        cmd, cwd=cwd, shell=True,
        stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
        creationflags=subprocess.CREATE_NEW_PROCESS_GROUP
    )
    PROCESSES.append(proc)
    print(f"  PID: {proc.pid}")

    if wait_seconds > 0:
        time.sleep(wait_seconds)

    print(f"  检测 port {port}...", end=" ")
    if wait_port(port, timeout=20):
        print("OK")
    else:
        print("TIMEOUT")
        return False

    if health_url:
        print(f"  健康检查 {health_url}...", end=" ")
        if check_health(health_url):
            print("OK")
        else:
            print("FAIL")
            return False

    print(f"  [{name}] 已就绪 OK")
    return True


def start_all_services() -> int:
    """启动所有 3 个本地服务，返回成功数"""
    successes = 0

    ok = start_service(
        cmd=[GO_BACKEND], cwd=GO_BACKEND_DIR,
        name="Go 后端", port=3000, wait_seconds=4,
    )
    if ok: successes += 1

    ok = start_service(
        cmd=["uvicorn", "api.app:app", "--host", "127.0.0.1", "--port", "8000", "--reload"],
        name="FastAPI", port=8000, cwd=BASE_DIR,
        health_url="http://127.0.0.1:8000/api/v1/health", wait_seconds=2,
    )
    if ok: successes += 1

    ok = start_service(
        cmd=["npm", "run", "dev"],
        name="Vue 前端", port=5173, cwd=FRONTEND_DIR, wait_seconds=6,
    )
    if ok: successes += 1

    return successes


# ── 状态显示 ──

def show_status():
    print("\n" + "=" * 50)
    print("  本地服务状态:")
    for name, port in [("Go 配送后端", 3000), ("FastAPI Agent", 8000), ("Vue 前端", 5173)]:
        s = socket.socket()
        s.settimeout(2)
        try:
            s.connect(("127.0.0.1", port))
            status = "running"
        except:
            status = "offline"
        finally:
            s.close()
        print(f"  {name:<16} (port {port})  {status}")

    # 检测 Docker 容器（直接查容器名，兼容不同 compose 来源）
    r = subprocess.run(
        ["docker", "ps", "--filter", "name=jd-agent", "-q"],
        capture_output=True, timeout=10
    )
    if r.returncode == 0 and r.stdout.strip():
        print("\n  Docker 容器:")
        docker_ps()
    print("=" * 50)


# ── CLI 入口 ──

def main():
    global DOCKER_RUNNING

    print("=" * 50)
    print("  JD-Agent 外卖智能助手")
    print("=" * 50)

    # 检查环境
    for cmd, name in [("python --version", "Python"), ("node --version", "Node.js"),
                      ("npm --version", "npm")]:
        r = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        if r.returncode != 0:
            print(f"[错误] 未找到 {name}")
            return 1
        print(f"  {name}: {r.stdout.strip()}")

    # 自动启动 Docker 基础设施（redis/postgres/rabbitmq）
    print("\n  [Docker] 检查基础设施...")
    docker_auto_start()

    successes = start_all_services()
    print(f"\n  启动结果: {successes}/3 服务成功")
    show_status()

    print(f"\n  Site: http://localhost:5173")
    print(f"  API:  http://localhost:8000/docs")


if __name__ == "__main__":
    # ── CLI 参数处理 ──
    if len(sys.argv) > 1:
        arg = sys.argv[1]
        if arg == "--start":
            sys.exit(main())
        elif arg == "--status":
            show_status()
            sys.exit(0)
        elif arg == "--stop":
            cleanup()
            print("所有服务已停止。")
            sys.exit(0)
        elif arg == "--docker":
            if not docker_check():
                sys.exit(1)
            r = subprocess.run(
                ["docker", "ps", "--filter", "name=jd-agent", "-q"],
                capture_output=True, timeout=10
            )
            if r.returncode == 0 and r.stdout.strip():
                docker_stop()
            else:
                if docker_up():
                    show_status()
            sys.exit(0)
        elif arg == "--start-service":
            if len(sys.argv) < 3:
                print("用法: python start_all.py --start-service <服务名>")
                print("服务名: go, fastapi, vue, docker")
                sys.exit(1)
            svc = sys.argv[2].lower()
            if svc == "docker":
                docker_auto_start()
            elif svc == "go":
                start_service(
                    cmd=[GO_BACKEND], cwd=GO_BACKEND_DIR,
                    name="Go 后端", port=3000, wait_seconds=4,
                )
            elif svc == "fastapi":
                start_service(
                    cmd=["uvicorn", "api.app:app", "--host", "127.0.0.1", "--port", "8000", "--reload"],
                    name="FastAPI", port=8000, cwd=BASE_DIR,
                    health_url="http://127.0.0.1:8000/api/v1/health", wait_seconds=2,
                )
            elif svc == "vue":
                start_service(
                    cmd=["npm", "run", "dev"],
                    name="Vue 前端", port=5173, cwd=FRONTEND_DIR, wait_seconds=6,
                )
            else:
                print(f"未知服务: {svc}")
                print("服务名: go, fastapi, vue, docker")
                sys.exit(1)
            print(f"\n  [{svc}] 已启动")
            sys.exit(0)
        elif arg == "--stop-service":
            if len(sys.argv) < 3:
                print("用法: python start_all.py --stop-service <服务名>")
                print("服务名: go, fastapi, vue, docker")
                sys.exit(1)
            svc = sys.argv[2].lower()
            if svc == "docker":
                docker_stop()
            elif svc == "go":
                kill_port(3000)
            elif svc == "fastapi":
                kill_port(8000)
            elif svc == "vue":
                kill_port(5173)
            else:
                print(f"未知服务: {svc}")
                sys.exit(1)
            print(f"\n  [{svc}] 已停止")
            sys.exit(0)
        elif arg == "--menu":
            # bat 模式：只做 docker 自动检测，不启动本地服务，bat 接管菜单
            print("\n  [Docker] 检查基础设施...")
            docker_auto_start()
            show_status()
            sys.exit(0)
        else:
            print(f"未知参数: {arg}")
            print(__doc__)
            sys.exit(1)

    # ── 交互模式 ──
    try:
        main()
    except KeyboardInterrupt:
        print("\n用户中断")
    except Exception as e:
        print(f"\n[错误] {e}")
    finally:
        cleanup()
