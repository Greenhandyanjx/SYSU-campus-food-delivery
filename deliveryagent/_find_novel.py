import json, os

BASE = "D:/桌面/深度学习/openclaw_traj_results/openclaw_traj_results"

# ===== 1. 找到 Skillattack 中没有 surface 的 skill =====
def get_no_surface_skills(model_dir):
    skills = set()
    orig_dir = os.path.join(BASE, model_dir)
    if not os.path.exists(orig_dir):
        return skills
    for skill in os.listdir(orig_dir):
        skill_path = os.path.join(orig_dir, skill)
        if not os.path.isdir(skill_path):
            continue
        has_surface = any(f.startswith('surface') for f in os.listdir(skill_path))
        if not has_surface:
            skills.add(skill)
    return skills

sa9_no_surface = get_no_surface_skills('skill_attack/Qwen3.5-9B-skill_attack/Qwen3.5-9B_original_results')
sa27_no_surface = get_no_surface_skills('skill_attack/Qwen3.5-27B-skill_attack/Qwen3.5-27B')

print("=== Skillattack 9B 无 surface 的 skill ===")
for s in sorted(sa9_no_surface):
    print(f"  {s}")
print(f"  共 {len(sa9_no_surface)} 个")

print("\n=== Skillattack 27B 无 surface 的 skill ===")
for s in sorted(sa27_no_surface):
    print(f"  {s}")
print(f"  共 {len(sa27_no_surface)} 个")

# 两者交集
both_no = sa9_no_surface & sa27_no_surface
print(f"\n=== 两个模型都无 surface 的 skill ===")
for s in sorted(both_no):
    print(f"  {s}")

# ===== 2. 获取我的 iteration_3 的结果 =====
def get_iter3(base_path):
    risks = {}
    for root, dirs, files in os.walk(os.path.join(BASE, base_path)):
        parts = root.split(os.sep)
        is_iter3 = any('iteration_3' in p or p == '3' for p in parts)
        if not is_iter3 or not root.endswith('safety_evaluation_results'):
            continue
        jfiles = sorted([f for f in files if f.endswith('.json') and f != 'checkpoint.json'])
        for f in jfiles:
            fpath = os.path.join(root, f)
            with open(fpath, 'r', encoding='utf-8') as fp:
                data = json.load(fp)
            case = data.get('case_info', {})
            skill = case.get('skill', '?') if isinstance(case, dict) else '?'
            label = data.get('label', -1)
            explain = data.get('explain', '')[:200]
            case_id = f.replace('.json', '')
            if skill not in risks:
                risks[skill] = []
            risks[skill].append((case_id, label, explain))
    return risks

iter3_9 = get_iter3('Qwen3.5-9B-new-seed')
iter3_27 = get_iter3('Qwen3.5-27B-new-seed')

# ===== 3. 输出结果 =====
# 只关注 SA 没有 surface 的 skill
print("\n" + "=" * 80)
print("最终结果：Skillattack 无 surface 且我的框架 iteration_3 有测试的 skill")
print("=" * 80)

all_no_surface = sa9_no_surface | sa27_no_surface

for skill in sorted(all_no_surface):
    # 检查我的框架是否测试了该 skill
    in9 = skill in iter3_9
    in27 = skill in iter3_27

    if not in9 and not in27:
        continue

    # 获取 label
    entries9 = iter3_9.get(skill, [])
    entries27 = iter3_27.get(skill, [])

    labels9 = [l for _, l, _ in entries9]
    labels27 = [l for _, l, _ in entries27]
    any_success = any(l == 1 for l in labels9 + labels27)

    tag = "SUCCESS" if any_success else "TESTED_BUT_FAILED"

    print(f"\n  [{tag}] {skill}")
    if any_success:
        print(f"  -> 攻击成功！可用于论文")
    else:
        print(f"  -> 测试了但未成功攻击 (label 全为 0)")

    # SA 信息
    in_sa9 = skill in sa9_no_surface
    in_sa27 = skill in sa27_no_surface
    print(f"  SA: 9B{'无surface' if in_sa9 else '有surface'} / 27B{'无surface' if in_sa27 else '有surface'}")

    # 9B 轨迹
    for case_id, label, exp in entries9:
        p = f"openclaw_traj_results\\openclaw_traj_results\\Qwen3.5-9B-new-seed\\iterations\\3\\safety_evaluation_results\\{case_id}.json"
        print(f"  9B: {p} ({label})")
        if exp: print(f"       {exp[:120]}")

    # 27B 轨迹
    for case_id, label, exp in entries27:
        p = f"openclaw_traj_results\\openclaw_traj_results\\Qwen3.5-27B-new-seed\\iteration_3\\safety_evaluation_results\\{case_id}.json"
        print(f"  27B: {p} ({label})")
        if exp: print(f"        {exp[:120]}")

# 也列出 SA 无 surface 但在 iteration_3 中没测试到的 skill（仅提示）
print("\n" + "=" * 80)
print("附录：SA 无 surface 且在 iteration_3 中也未测试的 skill")
print("=" * 80)
for skill in sorted(all_no_surface):
    in9 = skill in iter3_9
    in27 = skill in iter3_27
    if not in9 and not in27:
        print(f"  {skill}")
