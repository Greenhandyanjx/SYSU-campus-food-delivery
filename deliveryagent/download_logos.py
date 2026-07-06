"""Download brand logo images from the web."""
import urllib.request, os, sys

savedir = "D:/campus_food/images/merchant_logos"
os.makedirs(savedir, exist_ok=True)

# Known brand logo URLs
brands = {
    "mcdonalds": [
        "https://upload.wikimedia.org/wikipedia/commons/thumb/3/36/McDonald%27s_Golden_Arches.svg/200px-McDonald%27s_Golden_Arches.svg.png",
        "https://www.mcdonalds.com/content/dam/sites/usa/nfl/publication/logo-mcdonalds.png",
    ],
    "luckin_coffee": [
        "https://upload.wikimedia.org/wikipedia/zh/thumb/6/64/Luckin_coffee_logo.svg/200px-Luckin_coffee_logo.svg.png",
    ],
    "yidiandian": [],
    "yangguofu": [],
    "lanzhou_noodles": [],
    "xueyi_canteen": [],
    "shaxian_snakes": [],
    "zz_baozifan": [],
    "coco": [],
    "xinjiang_bbq": [],
    "shiweixuan_bento": [],
    "yakitori": [],
}

success = 0
for brand, urls in brands.items():
    ok = False
    for url in urls:
        try:
            req = urllib.request.Request(url, headers={"User-Agent": "Mozilla/5.0"})
            data = urllib.request.urlopen(req, timeout=10).read()
            ext = url.rsplit(".", 1)[-1].split("?")[0]
            if ext not in ("png", "jpg", "jpeg", "svg", "gif"):
                ext = "png"
            fpath = os.path.join(savedir, f"{brand}.{ext}")
            with open(fpath, "wb") as f:
                f.write(data)
            print(f"OK  {brand} -> {fpath} ({len(data)} bytes)")
            ok = True
            success += 1
            break
        except Exception as e:
            print(f"    {brand}: {e}")
    if not ok:
        print(f"MISS {brand}: no source")

print(f"\nDownloaded {success}/{len(brands)} logos")
