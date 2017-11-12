#ifndef SPELL_CACHE_H
#define SPELL_CACHE_H

#include <map>
#include <vector>

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
	struct sSkillProc {
		int chance;
		int skill;
		int spell;
	};

	SpellCache() : spell_cached(false), item_cached(false), alt_cached(false) {}
	~SpellCache() {}

	void InsertSpellEffect(int affect, int value, int subindex);
	void InsertItemEffect(int affect, int value, int subindex);
	void InsertAltEffect(int affect, int value, int subindex);
	void InsertSkillProc(int chace, int skill, int spell);

	inline void ClearSpellEffect() { m_spelleffect.clear(); }
	inline void ClearItemEffect() { m_itemeffect.clear(); }
	inline void ClearAltEffect() { m_alteffect.clear(); m_skill_proc.clear(); }

	inline void SetSpellCached(bool v) { spell_cached = v; }
	inline void SetItemCached(bool v) { item_cached = v; }
	inline void SetAltCached(bool v) { alt_cached = v; }

	inline bool IsSpellCached() { return spell_cached; }
	inline bool IsItemCached() { return item_cached; }
	inline bool IsAltCached() { return alt_cached; }
	inline bool HasSkillProcs() { return !m_skill_proc.empty(); }

	const sEffectCache *GetSpellCached(int affect, int subindex = 0);
	const sEffectCache *GetItemCached(int affect, int subindex = 0);
	const sEffectCache *GetAltCached(int affect, int subindex = 0);

	std::vector<sSkillProc>::const_iterator skill_proc_begin() { return m_skill_proc.cbegin(); }
	std::vector<sSkillProc>::const_iterator skill_proc_end() { return m_skill_proc.cend(); }

	// inlines for common operations
	inline int GetCachedPlayerEffect(int affect, int subindex = 0) {
		auto res = GetSpellCached(affect, subindex);
		if (res)
			return res->base1;
		return 0;
	}

	inline int GetCachedItemEffect(int affect, int subindex = 0) {
		auto res = GetItemCached(affect, subindex);
		if (res)
			return res->base1;
		return 0;
	}

	inline int GetCachedAltEffect(int affect, int subindex = 0) {
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
	std::vector<sSkillProc> m_skill_proc;
};

#endif /* !SPELL_CACHE_H */
