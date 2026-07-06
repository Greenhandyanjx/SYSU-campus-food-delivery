"""为所有商家生成品牌风格的 SVG logo 文件。"""
import os

SAVEDIR = "D:/campus_food/images/merchant_logos"
os.makedirs(SAVEDIR, exist_ok=True)

# 每个商家的品牌色和显示名称
BRANDS = [
    # (filename, brand_color, brand_name, icon_letter, accent_color)
    ("mcdonalds",       "#DA291C", "M", "M", "#FFC72C"),   # 麦当劳红
    ("luckin_coffee",   "#003366", "瑞", "幸", "#D4A574"),  # 瑞幸蓝
    ("yidiandian",      "#8BC34A", "一", "点", "#558B2F"),   # 一点点绿
    ("yangguofu",       "#E53935", "杨", "国", "#FF8A80"),   # 杨国福红
    ("lanzhou_noodles", "#FF8F00", "兰", "州", "#FFB300"),   # 兰州黄
    ("xueyi_canteen",   "#1565C0", "学", "一", "#42A5F5"),   # 学一蓝
    ("shaxian_snakes",  "#2E7D32", "沙", "县", "#66BB6A"),   # 沙县绿
    ("zz_baozifan",     "#6D4C41", "啫", "啫", "#A1887F"),   # 煲仔饭棕
    ("coco",            "#E65100", "c", "o", "#FF8A65"),     # coco橙
    ("xinjiang_bbq",    "#C62828", "夯", "肉", "#EF5350"),   # 新疆烧烤红
    ("shiweixuan_bento","#4A148C", "竹", "林", "#7E57C2"),   # 便当紫
    ("yakitori",        "#BF360C", "港", "都", "#FF7043"),   # 港都热炒橙红
    ("test_shop",       "#FFC107", "M", "", "#FFFFFF"),      # 测试（麦当劳备用）
]

def generate_svg(filename: str, bg: str, letter1: str, letter2: str, accent: str) -> str:
    """生成居中显示文字的圆角方形 SVG logo"""
    # 如果两个字符，显示成两行
    display = f"""<text x="100" y="95" text-anchor="middle" fill="white" font-size="60" font-weight="bold" font-family="Arial">{letter1}</text>"""
    if letter2:
        display += f"""\n<text x="100" y="145" text-anchor="middle" fill="white" font-size="36" font-weight="bold" font-family="Arial">{letter2}</text>"""

    return f'''<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200" viewBox="0 0 200 200">
  <defs>
    <linearGradient id="bg" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:{bg};stop-opacity:1" />
      <stop offset="100%" style="stop-color:{accent};stop-opacity:1" />
    </linearGradient>
  </defs>
  <rect width="200" height="200" fill="url(#bg)" rx="30"/>
  <rect x="10" y="10" width="180" height="180" rx="24" fill="none" stroke="rgba(255,255,255,0.2)" stroke-width="2"/>
  {display}
</svg>'''

count = 0
for filename, bg, l1, l2, accent in BRANDS:
    svg = generate_svg(filename, bg, l1, l2, accent)
    filepath = os.path.join(SAVEDIR, f"{filename}.svg")
    with open(filepath, "w", encoding="utf-8") as f:
        f.write(svg)
    print(f"  OK {filename}.svg")
    count += 1

print(f"\n生成 {count} 个 SVG logo 到 {SAVEDIR}")
