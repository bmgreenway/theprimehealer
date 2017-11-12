#include "spell_cache.h"

void SpellCache::InsertSpellEffect(int affect, int value, int subindex)
{
	auto range = m_spelleffect.equal_range(affect);
	auto effect_iter = range.first;

	while (effect_iter != range.second) {
		if (effect_iter->second.base2 == subindex)
			break;
		++effect_iter;
	}

	if (effect_iter == range.second) {// we didn't find one
		sEffectCache e{affect, value, subindex};
		m_spelleffect.emplace(affect, e);
		return;
	}

	// we gotta update
	effect_iter->second.base1 = value;
}

void SpellCache::InsertItemEffect(int affect, int value, int subindex)
{
	auto range = m_itemeffect.equal_range(affect);
	auto effect_iter = range.first;

	while (effect_iter != range.second) {
		if (effect_iter->second.base2 == subindex)
			break;
		++effect_iter;
	}

	if (effect_iter == range.second) {// we didn't find one
		sEffectCache e{affect, value, subindex};
		m_itemeffect.emplace(affect, e);
		return;
	}

	// we gotta update
	effect_iter->second.base1 = value;
}

void SpellCache::InsertAltEffect(int affect, int value, int subindex)
{
	auto range = m_alteffect.equal_range(affect);
	auto effect_iter = range.first;

	while (effect_iter != range.second) {
		if (effect_iter->second.base2 == subindex)
			break;
		++effect_iter;
	}

	if (effect_iter == range.second) {// we didn't find one
		sEffectCache e{affect, value, subindex};
		m_alteffect.emplace(affect, e);
		return;
	}

	// we gotta update
	effect_iter->second.base1 = value;
}

void SpellCache::InsertSkillProc(int chance, int skill, int spell)
{
	m_skill_proc.push_back({chance, skill, spell});
}

const SpellCache::sEffectCache *SpellCache::GetSpellCached(int affect, int subindex)
{
	auto range = m_spelleffect.equal_range(affect);
	auto effect_iter = range.first;

	while (effect_iter != range.second) {
		if (effect_iter->second.base2 == subindex)
			return &effect_iter->second;
		++effect_iter;
	}

	return nullptr;
}

const SpellCache::sEffectCache *SpellCache::GetItemCached(int affect, int subindex)
{
	auto range = m_itemeffect.equal_range(affect);
	auto effect_iter = range.first;

	while (effect_iter != range.second) {
		if (effect_iter->second.base2 == subindex)
			return &effect_iter->second;
		++effect_iter;
	}

	return nullptr;
}

const SpellCache::sEffectCache *SpellCache::GetAltCached(int affect, int subindex)
{
	auto range = m_alteffect.equal_range(affect);
	auto effect_iter = range.first;

	while (effect_iter != range.second) {
		if (effect_iter->second.base2 == subindex)
			return &effect_iter->second;
		++effect_iter;
	}

	return nullptr;
}

