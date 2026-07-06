"""Test ORDER_CARD extraction from agent response"""
import json, re, sys
from collections import Counter

with open("/tmp/agent_response.json", "r", encoding="utf-8") as f:
    raw = f.read()

try:
    data = json.loads(raw)
except Exception as e:
    print(f"JSON parse error: {e}")
    print("Raw preview:", raw[:300])
    sys.exit(1)

resp = data.get("response", "")
print("Response length:", len(resp))
print()

# Extract ORDER_CARD markers
cards = re.findall(r"\[ORDER_CARD_START\](.*?)\[ORDER_CARD_END\]", resp, re.DOTALL)
print(f"ORDER_CARD count: {len(cards)}")

for i, c in enumerate(cards):
    try:
        card = json.loads(c.strip())
        oid = card.get("orderId", "?")
        merch = card.get("merchant", "?")
        amt = card.get("amount", "?")
        st = card.get("status", "?")
        dishes = card.get("dishes", [])
        dish_names = [d.get("name","") for d in dishes]
        consignee = card.get("consignee", "")
        phone = card.get("phone", "")
        addr = card.get("address", "")
        print(f"  Card {i+1}: orderId={oid}, merchant={merch}, amount={amt}, status={st}")
        print(f"           dishes={dish_names}")
        print(f"           consignee={consignee}, phone={phone}, address={addr}")
    except Exception as e:
        print(f"  Card {i+1}: PARSE ERROR: {e}")
        print(f"  Raw: {c[:200]}")

# Dedup check
order_ids = []
for c in cards:
    try:
        card = json.loads(c.strip())
        order_ids.append(card.get("orderId"))
    except:
        pass
dupes = {k:v for k,v in Counter(order_ids).items() if v > 1}
if dupes:
    print(f"\n  ⚠️ DUPLICATE orderIds: {dupes}")
else:
    print("\n  ✅ No duplicate orderIds")

print("\n=== FULL RESPONSE ===")
print(resp)
