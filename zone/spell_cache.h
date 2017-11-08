#ifndef SPELL_CACHE_H
#define SPELL_CACHE_H

#include <map>

class SpellCache
{
public:
	struct sEffectCache {
		int affect;
		int base1;
		int base2;
	};

	SpellCache() : spell_cached(false), item_cached(false), alt_cached(false) {}
	~SpellCache() {}

	void InsertSpellEffect(int affect, int value, int subindex);
	void InsertItemEffect(int affect, int value, int subindex);
	void InsertAltEffect(int affect, int value, int subindex);

	inline void ClearSpellEffect() { m_spelleffect.clear(); }
	inline void ClearItemEffect() { m_itemeffect.clear(); }
	inline void ClearAltEffect() { m_alteffect.clear(); }

	inline void SetSpellCached(bool v) { spell_cached = v; }
	inline void SetItemCached(bool v) { item_cached = v; }
	inline void SetAltCached(bool v) { alt_cached = v; }

	inline bool IsSpellCached() { return spell_cached; }
	inline bool IsItemCached() { return item_cached; }
	inline bool IsAltCached() { return alt_cached; }

	const sEffectCache *GetSpellCached(int affect, int subindex = 0);
	const sEffectCache *GetItemCached(int affect, int subindex = 0);
	const sEffectCache *GetAltCached(int affect, int subindex = 0);

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
};

#endif /* !SPELL_CACHE_H */
