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

void SpellCache::InsertSkillAttackProc(int chance, int skill, int spell)
{
	m_skill_attack_proc.push_back({chance, skill, spell});
}

const SpellCache::sEffectCache *SpellCache::GetSpellCached(int affect, int subindex) const
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

const SpellCache::sEffectCache *SpellCache::GetItemCached(int affect, int subindex) const
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

const SpellCache::sEffectCache *SpellCache::GetAltCached(int affect, int subindex) const
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

void SpellCache::InsertSkillProcAttempt(SpellCache::eSkillProc type, int spell, int chance, int slot)
{
	sSkillProc e;
	e.spell = spell;
	e.chance = chance;
	e.slot = slot;

	switch (type) {
	case eSkillProc::AA:
		m_skill_proc_attempt.aa.push_back(e);
		return;
	case eSkillProc::Buff:
		m_skill_proc_attempt.buff.push_back(e);
		return;
	case eSkillProc::Worn:
		m_skill_proc_attempt.worn.push_back(e);
		return;
	}
}

void SpellCache::InsertSkillProcSuccess(SpellCache::eSkillProc type, int spell, int chance, int slot)
{
	sSkillProc e;
	e.spell = spell;
	e.chance = chance;
	e.slot = slot;

	switch (type) {
	case eSkillProc::AA:
		m_skill_proc_success.aa.push_back(e);
		return;
	case eSkillProc::Buff:
		m_skill_proc_success.buff.push_back(e);
		return;
	case eSkillProc::Worn:
		m_skill_proc_success.worn.push_back(e);
		return;
	}
}

void SpellCache::InsertSkillLimit(SpellCache::eSkillProc type, EQEmu::skills::SkillType skill, bool on_success)
{
	auto &which = on_success ? m_skill_proc_success : m_skill_proc_attempt;

	// we have to assume the last one inserted is correct, otherwise it would be bad data
	switch (type) {
	case eSkillProc::AA: {
		auto it = which.aa.rbegin();
		if (it != which.aa.rend())
			it->skills.insert(skill);
		return;
	}
	case eSkillProc::Buff: {
		auto it = which.buff.rbegin();
		if (it != which.buff.rend())
			it->skills.insert(skill);
		return;
	}
	case eSkillProc::Worn: {
		auto it = which.worn.rbegin();
		if (it != which.worn.rend())
			it->skills.insert(skill);
		return;
	}
	}
}

std::vector<SpellCache::sSkillProc>::const_iterator SpellCache::skill_proc_attempt_begin(SpellCache::eSkillProc type) const
{
	switch (type) {
	case eSkillProc::AA:
		return m_skill_proc_attempt.aa.cbegin();
	case eSkillProc::Buff:
		return m_skill_proc_attempt.buff.cbegin();
	case eSkillProc::Worn:
		return m_skill_proc_attempt.worn.cbegin();
	}
}

std::vector<SpellCache::sSkillProc>::const_iterator SpellCache::skill_proc_attempt_end(SpellCache::eSkillProc type) const
{
	switch (type) {
	case eSkillProc::AA:
		return m_skill_proc_attempt.aa.cend();
	case eSkillProc::Buff:
		return m_skill_proc_attempt.buff.cend();
	case eSkillProc::Worn:
		return m_skill_proc_attempt.worn.cend();
	}
}

std::vector<SpellCache::sSkillProc>::const_iterator SpellCache::skill_proc_success_begin(SpellCache::eSkillProc type) const
{
	switch (type) {
	case eSkillProc::AA:
		return m_skill_proc_success.aa.cbegin();
	case eSkillProc::Buff:
		return m_skill_proc_success.buff.cbegin();
	case eSkillProc::Worn:
		return m_skill_proc_success.worn.cbegin();
	}
}

std::vector<SpellCache::sSkillProc>::const_iterator SpellCache::skill_proc_success_end(SpellCache::eSkillProc type) const
{
	switch (type) {
	case eSkillProc::AA:
		return m_skill_proc_success.aa.cend();
	case eSkillProc::Buff:
		return m_skill_proc_success.buff.cend();
	case eSkillProc::Worn:
		return m_skill_proc_success.worn.cend();
	}
}

