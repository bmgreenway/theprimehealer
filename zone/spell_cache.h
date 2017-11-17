#ifndef SPELL_CACHE_H
#define SPELL_CACHE_H

#include <map>
#include <vector>
#include <set>

#include "../common/skills.h"

class SpellCache
{
public:
	struct sEffectCache {
		int affect;
		int base1;
		int base2;
	};

	// can't really get away with doing a cache trick with this since you can have multiple
	// (see Decapitation)
	// This is for SE_SkillAttackProc (SPA 288) only, which is purely AA
	struct sSkillAttackProc {
		int chance;
		int skill;
		int spell;
	};

	// enum for skill proc functions
	enum class eSkillProc {
		Buff,
		Worn,
		AA
	};

	// SE_SpellOnDeath
	struct sDeathProc {
		int chance;
		int spell;
		int slot;
	};

	// this is used for SE_SkillProc and SE_SkillProcSuccess which are matched with SE_LimitToSkill
	struct sSkillProc {
		int chance;
		int spell;
		int slot; // used for buff slot when it's a buff
		std::set<EQEmu::skills::SkillType> skills;
	};

	// we need to recache these separately
	template <typename T>
	struct sProcs {
		std::vector<T> buff;
		std::vector<T> worn;
		std::vector<T> aa;
	};

	SpellCache() : spell_cached(false), item_cached(false), alt_cached(false) {}
	~SpellCache() {}

	void InsertSpellEffect(int affect, int value, int subindex);
	void InsertItemEffect(int affect, int value, int subindex);
	void InsertAltEffect(int affect, int value, int subindex);
	void InsertSkillAttackProc(int chace, int skill, int spell);

	inline void ClearSpellEffect() { m_spelleffect.clear(); m_skill_proc_attempt.buff.clear(); m_skill_proc_success.buff.clear(); m_death_proc.buff.clear(); }
	inline void ClearItemEffect() { m_itemeffect.clear(); m_skill_proc_attempt.worn.clear(); m_skill_proc_success.worn.clear(); m_death_proc.worn.clear(); }
	inline void ClearAltEffect() { m_alteffect.clear(); m_skill_attack_proc.clear(); m_skill_proc_attempt.aa.clear(); m_skill_proc_success.aa.clear(); m_death_proc.aa.clear(); }

	inline void SetSpellCached(bool v) { spell_cached = v; }
	inline void SetItemCached(bool v) { item_cached = v; }
	inline void SetAltCached(bool v) { alt_cached = v; }

	inline bool IsSpellCached() const { return spell_cached; }
	inline bool IsItemCached() const { return item_cached; }
	inline bool IsAltCached() const { return alt_cached; }
	inline bool HasSkillAttackProcs() const { return !m_skill_attack_proc.empty(); }

	void InsertSkillProcAttempt(eSkillProc type, int spell, int chance, int slot = -1);
	void InsertSkillProcSuccess(eSkillProc type, int spell, int chance, int slot = -1);
	void InsertSkillLimit(eSkillProc type, EQEmu::skills::SkillType skill, bool on_success);
	inline bool HasSkillProcAttempt() const { return !m_skill_proc_attempt.aa.empty() || !m_skill_proc_attempt.buff.empty() || !m_skill_proc_attempt.worn.empty(); }
	inline bool HasSkillProcSuccess() const { return !m_skill_proc_success.aa.empty() || !m_skill_proc_success.buff.empty() || !m_skill_proc_success.worn.empty(); }

	void InsertDeathProc(eSkillProc type, int spell, int chance, int slot = -1);

	const sEffectCache *GetSpellCached(int affect, int subindex = 0) const;
	const sEffectCache *GetItemCached(int affect, int subindex = 0) const;
	const sEffectCache *GetAltCached(int affect, int subindex = 0) const;

	std::vector<sSkillAttackProc>::const_iterator skill_attack_proc_begin() const { return m_skill_attack_proc.cbegin(); }
	std::vector<sSkillAttackProc>::const_iterator skill_attack_proc_end() const { return m_skill_attack_proc.cend(); }

	std::vector<sSkillProc>::const_iterator skill_proc_attempt_begin(eSkillProc type) const;
	std::vector<sSkillProc>::const_iterator skill_proc_attempt_end(eSkillProc type) const;

	std::vector<sSkillProc>::const_iterator skill_proc_success_begin(eSkillProc type) const;
	std::vector<sSkillProc>::const_iterator skill_proc_success_end(eSkillProc type) const;

	std::vector<sDeathProc>::const_iterator death_proc_begin(eSkillProc type) const;
	std::vector<sDeathProc>::const_iterator death_proc_end(eSkillProc type) const;

	// inlines for common operations
	inline int GetCachedPlayerEffect(int affect, int subindex = 0) const {
		auto res = GetSpellCached(affect, subindex);
		if (res)
			return res->base1;
		return 0;
	}

	inline int GetCachedItemEffect(int affect, int subindex = 0) const {
		auto res = GetItemCached(affect, subindex);
		if (res)
			return res->base1;
		return 0;
	}

	inline int GetCachedAltEffect(int affect, int subindex = 0) const {
		auto res = GetAltCached(affect, subindex);
		if (res)
			return res->base1;
		return 0;
	}

private:
	bool spell_cached;
	bool item_cached;
	bool alt_cached;
	// maybe use unordered_multimap
	std::multimap<int, sEffectCache> m_spelleffect;
	std::multimap<int, sEffectCache> m_itemeffect;
	std::multimap<int, sEffectCache> m_alteffect;

	// this should be fine for how it works with live AAs
	// if custom servers want to give someone a million of these ahhh
	// we need something better :P
	std::vector<sSkillAttackProc> m_skill_attack_proc;

	sProcs<sSkillProc> m_skill_proc_attempt; // SE_SkillProc
	sProcs<sSkillProc> m_skill_proc_success; // SE_SkillProcSuccess
	sProcs<sDeathProc> m_death_proc; // SE_SpellOnDeath
};

#endif /* !SPELL_CACHE_H */
