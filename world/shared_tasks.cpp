#include "../common/string_util.h"
#include "cliententry.h"
#include "clientlist.h"
#include "shared_tasks.h"
#include "worlddb.h"
#include "zonelist.h"

#include <algorithm>
#include <fmt/format.h>

extern ClientList client_list;
extern ZSList zoneserver_list;
extern SharedTaskManager shared_tasks;

void SharedTaskManager::HandleTaskRequest(ServerPacket *pack)
{
	if (!pack)
		return;

	/*
	 * Things done in zone:
	 * Verified we were requesting a shared task
	 * Verified leader has a slot available (guess we should double check this one)
	 * Verified leader met level reqs
	 * Verified repeatable or not completed (not doing that here?)
	 * Verified leader doesn't have a lock out
	 * Verified the group/raid met min/max player counts
	 */

	char tmp_str[64] = { 0 };
	int task_id = pack->ReadUInt32();
	int npc_id = pack->ReadUInt32();
	pack->ReadString(tmp_str);
	std::string leader_name = tmp_str;
	int player_count = pack->ReadUInt32();
	std::vector<std::string> players;
	for (int i = 0; i < player_count; ++i) {
		pack->ReadString(tmp_str);
		players.push_back(tmp_str);
	}

	// check if the task exist, we only load shared tasks in world, so we know the type is correct if found
	auto it = task_information.find(task_id);
	if (it == task_information.end()) { // not loaded! bad id or not shared task
		auto pc = client_list.FindCharacter(leader_name.c_str());
		if (pc) {
			// failure TODO: appropriate message
			auto pack = new ServerPacket(ServerOP_TaskReject, leader_name.size() + 1 + 8);
			pack->WriteUInt32(0); // string ID or just generic fail message
			pack->WriteUInt32(npc_id);
			pack->WriteString(leader_name.c_str());
			zoneserver_list.SendPacket(pc->zone(), pc->instance(), pack);
			safe_delete(pack);
		} // oh well
		return;
	}

	int id = GetNextID();
	auto ret = tasks.insert({id, {id, task_id}});
	if (!ret.second) {
		auto pc = client_list.FindCharacter(leader_name.c_str());
		if (pc) {
			// failure TODO: appropriate message
			auto pack = new ServerPacket(ServerOP_TaskReject, leader_name.size() + 1 + 8);
			pack->WriteUInt32(0); // string ID or just generic fail message
			pack->WriteUInt32(npc_id);
			pack->WriteString(leader_name.c_str());
			zoneserver_list.SendPacket(pc->zone(), pc->instance(), pack);
			safe_delete(pack);
		} // oh well
		return;
	}

	auto cle_leader = client_list.FindCharacter(leader_name.c_str());
	if (cle_leader == nullptr) {// something went wrong
		tasks.erase(ret.first);
		return;
	}

	if (!cle_leader->HasFreeSharedTaskSlot()) { // they have a task already ...
		tasks.erase(ret.first);
		return;
	}

	auto &task = ret.first->second;
	task.AddMember(leader_name, cle_leader, cle_leader->CharID(), true);

	if (players.empty()) {
		// send instant success to leader
		SerializeBuffer buf(10);
		buf.WriteInt32(id);				// shared task's ID
		buf.WriteInt32(task_id);		// ID of the task's data
		buf.WriteInt32(npc_id);			// NPC we're requesting from
		buf.WriteString(leader_name);	// leader's name
		buf.WriteInt32(0); // member list minus leader

		auto pack = new ServerPacket(ServerOP_TaskGrant, buf);
		zoneserver_list.SendPacket(cle_leader->zone(), cle_leader->instance(), pack);
		safe_delete(pack);

		task.SetCLESharedTasks();
		return;
	}

	for (auto &&name : players) {
		// look up CLEs by name, tell them we need to know if they can be added
		auto cle = client_list.FindCharacter(name.c_str());
		if (cle) {
			// make sure we don't have a shared task already
			if (!cle->HasFreeSharedTaskSlot()) {
				// failure TODO: appropriate message
				auto pack = new ServerPacket(ServerOP_TaskReject, leader_name.size() + 1 + 8);
				pack->WriteUInt32(0); // string ID or just generic fail message
				pack->WriteUInt32(npc_id);
				pack->WriteString(leader_name.c_str());
				zoneserver_list.SendPacket(cle_leader->zone(), cle_leader->instance(), pack);
				safe_delete(pack);
				tasks.erase(ret.first);
				return;
			}

			// make sure our level is right
			if (!AppropriateLevel(task_id, cle->level())) {
				// failure TODO: appropriate message
				auto pack = new ServerPacket(ServerOP_TaskReject, leader_name.size() + 1 + 8);
				pack->WriteUInt32(0); // string ID or just generic fail message
				pack->WriteUInt32(npc_id);
				pack->WriteString(leader_name.c_str());
				zoneserver_list.SendPacket(cle_leader->zone(), cle_leader->instance(), pack);
				safe_delete(pack);
				tasks.erase(ret.first);
				return;
			}

			// check our lock out timer
			int expires = cle->GetTaskLockoutExpire(task_id);
			if ((expires - time(nullptr)) >= 0) {
				// failure TODO: appropriate message, we need to send the timestamp here
				auto pack = new ServerPacket(ServerOP_TaskReject, leader_name.size() + 1 + 8);
				pack->WriteUInt32(0); // string ID or just generic fail message
				pack->WriteUInt32(npc_id);
				pack->WriteString(leader_name.c_str());
				zoneserver_list.SendPacket(cle_leader->zone(), cle_leader->instance(), pack);
				safe_delete(pack);
				tasks.erase(ret.first);
				return;
			}

			// we're good, add to task
			task.AddMember(name, cle, cle->CharID());
		}
	}

	// this will also prevent any of these clients from requesting or being added to another, lets do it now before we tell zone
	task.SetCLESharedTasks();
	task.InitActivities();
	task.SetUpdated();
	// fire off to zone we're done!
	SerializeBuffer buf(10 + 10 * players.size());
	buf.WriteInt32(id);				// shared task's ID
	buf.WriteInt32(task_id);		// ID of the task's data
	buf.WriteInt32(npc_id);			// NPC we're requesting from
	buf.WriteInt32(task.GetAcceptedTime());	// time we accepted it
	buf.WriteString(leader_name);	// leader's name
	task.SerializeMembers(buf, false);	// everyone but leader

	auto reply = new ServerPacket(ServerOP_TaskGrant, buf);
	zoneserver_list.SendPacket(cle_leader->zone(), cle_leader->instance(), reply);
	safe_delete(reply);

	task.Save();

	return;
}

/*
 * Just sends the ID of the task that was successfully created zone side
 * We now need to tell all the other clients to join the task
 * We could probably try to find all the clients already in the zone and not
 * worry about them here, but it's simpler this way
 */
void SharedTaskManager::HandleTaskZoneCreated(ServerPacket *pack)
{
	if (!pack)
		return;

	int id = pack->ReadUInt32();

	auto task = GetSharedTask(id);

	if (!task) // hmm guess we should tell zone something is broken TODO
		return;

	// we reuse this, easier this way
	auto outpack = new ServerPacket(ServerOP_TaskZoneCreated, sizeof(ServerSharedTaskMember_Struct));
	auto stm = (ServerSharedTaskMember_Struct *)outpack->pBuffer;
	stm->id = id;

	for (auto &&m : task->members.list) {
		if (m.leader) // leader done!
			continue;

		if (!m.cle) // hmmm
			continue;

		if (!m.cle->Server()) // hmm
			continue;

		strn0cpy(stm->name, m.name.c_str(), 64);
		zoneserver_list.SendPacket(m.cle->zone(), m.cle->instance(), outpack);
	}

	safe_delete(outpack);
}

void SharedTaskManager::HandleTaskActivityUpdate(ServerPacket *pack)
{
	if (!pack)
		return;

	if (pack->size != sizeof(ServerOP_TaskActivityUpdate))
		return;

	auto update = (ServerSharedTaskActivity_Struct *)pack->pBuffer;

	auto task = GetSharedTask(update->id);

	if (!task) // guess it wasn't loaded?
		return;

	task->ProcessActivityUpdate(update->activity_id, update->value);
}

/*
 * Loads in the tasks and task_activity tables
 * We limit to shared to save some memory
 * This can be called while reloading tasks (because deving etc)
 * This data is loaded into the task_information map
 */

bool SharedTaskManager::LoadSharedTasks(int single_task)
{
	std::string query;

	if (single_task == 0) {
		query =
		    StringFormat("SELECT `id`, `type`, `duration`, `duration_code`, `title`, `description`, `reward`, "
				 "`rewardid`, `cashreward`, `xpreward`, `rewardmethod`, `faction_reward`, `minlevel`, "
				 "`maxlevel`, `repeatable`, `completion_emote`, `reward_points`, `reward_type`, "
				 "`replay_group`, `min_players`, `max_players`, `task_lock_step`, `instance_zone_id`, "
				 "`zone_version`, `zone_in_zone_id`, `zone_in_x`, `zone_in_y`, `zone_in_object_id`, "
				 "`dest_x`, `dest_y`, `dest_z`, `dest_h` FROM `tasks` WHERE `type` = %i",
				 static_cast<int>(TaskType::Shared));
	} else {
		query =
		    StringFormat("SELECT `id`, `type`, `duration`, `duration_code`, `title`, `description`, `reward`, "
				 "`rewardid`, `cashreward`, `xpreward`, `rewardmethod`, `faction_reward`, `minlevel`, "
				 "`maxlevel`, `repeatable`, `completion_emote`, `reward_points`, `reward_type`, "
				 "`replay_group`, `min_players`, `max_players`, `task_lock_step`, `instance_zone_id`, "
				 "`zone_version`, `zone_in_zone_id`, `zone_in_x`, `zone_in_y`, `zone_in_object_id`, "
				 "`dest_x`, `dest_y`, `dest_z`, `dest_h` FROM `tasks` WHERE `id` = %i AND `type` = %i",
				 single_task, static_cast<int>(TaskType::Shared));
	}
	auto results = database.QueryDatabase(query);
	if (!results.Success()) {
		return false;
	}

	for (auto row = results.begin(); row != results.end(); ++row) {
		int task_id = atoi(row[0]);

		auto &task = task_information[task_id];
		task.type = static_cast<TaskType>(atoi(row[1]));
		task.Duration = atoi(row[2]);
		task.dur_code = static_cast<DurationCode>(atoi(row[3]));
		task.Title = row[4];
		task.Description = row[5];
		task.Reward = row[6];
		task.RewardID = atoi(row[7]);
		task.CashReward = atoi(row[8]);
		task.XPReward = atoi(row[9]);
		task.RewardMethod = (TaskMethodType)atoi(row[10]);
		task.faction_reward = atoi(row[11]);
		task.MinLevel = atoi(row[12]);
		task.MaxLevel = atoi(row[13]);
		task.Repeatable = atoi(row[14]);
		task.completion_emote = row[15];
		task.reward_points = atoi(row[16]);
		task.reward_type = static_cast<PointType>(atoi(row[17]));
		task.replay_group = atoi(row[18]);
		task.min_players = atoi(row[19]);
		task.max_players = atoi(row[20]);
		task.task_lock_step = atoi(row[21]);
		task.instance_zone_id = atoi(row[22]);
		task.zone_version = atoi(row[23]);
		task.zone_in_zone_id = atoi(row[24]);
		task.zone_in_x = atof(row[25]);
		task.zone_in_y = atof(row[26]);
		task.zone_in_object_id = atoi(row[27]);
		task.dest_x = atof(row[28]);
		task.dest_y = atof(row[29]);
		task.dest_z = atof(row[30]);
		task.dest_h = atof(row[31]);
		task.ActivityCount = 0;
		task.SequenceMode = ActivitiesSequential;
		task.LastStep = 0;
	}

	// hmm need to limit to shared tasks only ...
	if (single_task == 0)
		query = StringFormat(
		    "SELECT `taskid`, `step`, `activityid`, `activitytype`, `target_name`, `item_list`, `skill_list`, "
		    "`spell_list`, `description_override`, `goalid`, `goalmethod`, `goalcount`, `delivertonpc`, "
		    "`zones`, `optional` FROM `task_activities` WHERE `activityid` < %i AND `taskid` IN (SELECT `id` "
		    "FROM `tasks` WHERE `type` = %i) ORDER BY taskid, activityid ASC",
		    MAXACTIVITIESPERTASK, static_cast<int>(TaskType::Shared));
	else
		query = StringFormat(
		    "SELECT `taskid`, `step`, `activityid`, `activitytype`, `target_name`, `item_list`, `skill_list`, "
		    "`spell_list`, `description_override`, `goalid`, `goalmethod`, `goalcount`, `delivertonpc`, "
		    "`zones`, `optional` FROM `task_activities` WHERE `taskid` = %i AND `activityid` < %i AND `taskid` "
		    "IN (SELECT `id` FROM `tasks` WHERE `type` = %i) ORDER BY taskid, activityid ASC",
		    single_task, MAXACTIVITIESPERTASK, static_cast<int>(TaskType::Shared));
	results = database.QueryDatabase(query);
	if (!results.Success()) {
		return false;
	}

	for (auto row = results.begin(); row != results.end(); ++row) {
		int task_id = atoi(row[0]);
		int step = atoi(row[1]);

		int activity_id = atoi(row[2]);

		if (activity_id < 0 || activity_id >= MAXACTIVITIESPERTASK) {
			// This shouldn't happen, as the SELECT is bounded by MAXTASKS
			continue;
		}

		if (task_information.count(task_id) == 0) {
			continue;
		}

		auto &task = task_information[task_id];

		task.Activity[task.ActivityCount].StepNumber = step;

		if (step != 0)
			task.SequenceMode = ActivitiesStepped;

		if (step > task.LastStep)
			task.LastStep = step;

		// Task Activities MUST be numbered sequentially from 0. If not, log an error
		// and set the task to nullptr. Subsequent activities for this task will raise
		// ERR_NOTASK errors.
		// Change to (activityID != (task.ActivityCount + 1)) to index from 1
		if (activity_id != task.ActivityCount) {
			task_information.erase(task_id);
			continue;
		}

		task.Activity[task.ActivityCount].Type = atoi(row[3]);

		task.Activity[task.ActivityCount].target_name = row[4];
		task.Activity[task.ActivityCount].item_list = row[5];
		task.Activity[task.ActivityCount].skill_list = row[6];
		task.Activity[task.ActivityCount].skill_id = atoi(row[6]); // for older clients
		task.Activity[task.ActivityCount].spell_list = row[7];
		task.Activity[task.ActivityCount].spell_id = atoi(row[7]); // for older clients
		task.Activity[task.ActivityCount].desc_override = row[8];

		task.Activity[task.ActivityCount].GoalID = atoi(row[9]);
		task.Activity[task.ActivityCount].GoalMethod = (TaskMethodType)atoi(row[10]);
		task.Activity[task.ActivityCount].GoalCount = atoi(row[11]);
		task.Activity[task.ActivityCount].DeliverToNPC = atoi(row[12]);
		task.Activity[task.ActivityCount].zones = row[13];
		auto zones = SplitString(task.Activity[task.ActivityCount].zones, ';');
		for (auto && e : zones)
			task.Activity[task.ActivityCount].ZoneIDs.push_back(std::stoi(e));
		task.Activity[task.ActivityCount].Optional = atoi(row[14]);

		task.ActivityCount++;
	}

	return true;
}

/*
 * This is called once during boot of world
 * We need to load next_id, clean up expired tasks (?), and populate the map
 */
bool SharedTaskManager::LoadSharedTaskState()
{
	// one may think we should clean up expired tasks, but we don't just in case world is booting back up after a crash
	// we will clean them up in the normal process loop so zones get told to clean up
	std::string query =
	    "SELECT `id`, `task_id`, `accepted_time`, `is_locked`, `is_completed` FROM `shared_task_state`";
	auto results = database.QueryDatabase(query);

	if (results.Success() && results.RowCount() > 0) {
		for (auto row = results.begin(); row != results.end(); ++row) {
			int id = atoi(row[0]);

			auto &task = tasks[id];
			task.task_state.slot = 0;
			task.task_state.Updated = false;
			task.SetID(id);
			task.SetTaskID(atoi(row[1]));
			task.SetAcceptedTime(atoi(row[2]));
			task.SetLocked(atoi(row[3]) != 0);
			task.SetCompleted(atoi(row[4]) != 0);
		}
	}

	query = "SELECT `shared_task_id`, `character_id`, `character_name`, `is_leader` FROM `shared_task_members` ORDER BY shared_task_id ASC";
	results = database.QueryDatabase(query);
	if (results.Success() && results.RowCount() > 0) {
		for (auto row = results.begin(); row != results.end(); ++row) {
			int task_id = atoi(row[0]);
			// hmm not sure best way to do this, fine for now
			if (tasks.count(task_id) == 1)
				tasks[task_id].AddMember(row[2], nullptr, atoi(row[1]), atoi(row[3]) != 0);
		}
	}

	query = "SELECT `shared_task_id`, `activity_id`, `done_count`, `completed` FROM `shared_task_activities` ORDER BY shared_task_id ASC";
	results = database.QueryDatabase(query);
	if (results.Success() && results.RowCount() > 0) {
		for (auto row = results.begin(); row != results.end(); ++row) {
			int task_id = atoi(row[0]);
			// hmm not sure best way to do this, fine for now
			if (tasks.count(task_id) == 1) {
				int index = atoi(row[1]);
				auto &task = tasks[task_id];
				task.task_state.Activity[index].ActivityID = index;
				task.task_state.Activity[index].DoneCount = atoi(row[2]);
				if (atoi(row[3]) != 0)
					task.task_state.Activity[index].State = ActivityCompleted;
				else
					task.task_state.Activity[index].State = ActivityHidden;
			}
		}
	}

	// TODO we need to call UnlockActivities on all the tasks ....

	// Load existing tasks. We may not want to actually do this here and wait for a client to log in
	// But the crash case may actually dictate we should :P

	// set next_id to highest used ID
	query = "SELECT IFNULL(MAX(id), 0) FROM shared_task_state";
	results = database.QueryDatabase(query);
	if (results.Success() && results.RowCount() == 1) {
		auto row = results.begin();
		next_id = atoi(row[0]);
	} else {
		next_id = 0; // oh well
	}

	return true;
}

/*
 * Return the next unused ID
 * Hopefully this does not grow too large.
 */
int SharedTaskManager::GetNextID()
{
	next_id++;
	// let's not be extra clever here ...
	while (tasks.count(next_id) != 0)
		next_id++;

	return next_id;
}

/*
 * returns true if the level fits in the task's defined range
 */
bool SharedTaskManager::AppropriateLevel(int id, int level) const
{
	auto it = task_information.find(id);
	// doesn't exist
	if (it == task_information.end())
		return false;

	auto &task = it->second;

	if (task.MinLevel && level < task.MinLevel)
		return false;

	if (task.MaxLevel && level > task.MaxLevel)
		return false;

	return true;
}

/*
 * This will check if any tasks have expired
 */
void SharedTaskManager::Process()
{
}

/*
 * When a player leaves world they will tell us to clean up their pointer
 * This is NOT leaving the shared task, just crashed or something
 */

void SharedTask::MemberLeftGame(ClientListEntry *cle)
{
	auto it = std::find_if(members.list.begin(), members.list.end(), [cle](SharedTaskMember &m) { return m.cle == cle; });

	// ahh okay ...
	if (it == members.list.end())
		return;

	it->cle = nullptr;
}

/*
 * Serializes Members into the SerializeBuffer
 * Starts with count then followed by names null-termed
 * In the future this will include monster mission shit
 * This should probably send the SharedMember struct or something more like it, fine for now
 */
void SharedTask::SerializeMembers(SerializeBuffer &buf, bool include_leader) const
{
	buf.WriteInt32(include_leader ? members.list.size() : members.list.size() - 1);

	for (auto && m : members.list) {
		if (!include_leader && m.leader)
			continue;

		buf.WriteString(m.name);
		// TODO: live also has monster mission class choice in here
	}
}

/*
 * This sets the CLE's quick look up shared task stuff
 */
void SharedTask::SetCLESharedTasks()
{
	for (auto &&m : members.list) {
		if (m.cle == nullptr) // shouldn't happen ....
			continue;

		m.cle->SetSharedTask(this);
		m.cle->SetCurrentSharedTaskID(id);
	}
}

void SharedTask::Save()
{
	const char *ERR_MYSQLERROR = "[TASKS]Error in TaskManager::SaveClientState %s";
	database.TransactionBegin();
	std::string query; // simple queries
	fmt::basic_memory_buffer<char> out; // queries where we loop over stuff

	if (task_state.Updated) {
		query = fmt::format("REPLACE INTO shared_task_state (id, task_id, accepted_time, is_locked, "
				    "is_completed) VALUES ({}, {}, {}, {:d}. {:d})",
				    id, task_id, GetAcceptedTime(), locked, completed);
		auto res = database.QueryDatabase(query);
		if (!res.Success())
			Log(Logs::General, Logs::Error, ERR_MYSQLERROR, res.ErrorMessage().c_str());
		else
			task_state.Updated = false;
	}

	int activity_count = 0;
	int max = shared_tasks.GetTaskActivityCount(task_id);
	fmt::format_to(
	    out, "REPLACE INTO shared_task_activities (shared_task_id, activity_id, done_count, completed) VALUES ");
	for (int i = 0; i < max; ++i) {
		if (!task_state.Activity[i].Updated)
			continue;

		if (activity_count == 0)
			fmt::format_to(out, "({}, {}, {}, {:d})", id, i, task_state.Activity[i].DoneCount,
				       task_state.Activity[i].State == ActivityCompleted);
		else
			fmt::format_to(out, ", ({}, {}, {}, {:d})", id, i, task_state.Activity[i].DoneCount,
				       task_state.Activity[i].State == ActivityCompleted);
		++activity_count;
	}

	// we got stuff to write
	if (activity_count != 0) {
		query = fmt::to_string(out);
		out.clear();
		auto res = database.QueryDatabase(query);
		if (!res.Success()) {
			Log(Logs::General, Logs::Error, ERR_MYSQLERROR, res.ErrorMessage().c_str());
		} else {
			for (int i = 0; i < max; ++i)
				task_state.Activity[i].Updated = false;
		}
	}

	if (members.update) {
		query = fmt::format("DELETE FROM `shared_task_members` WHERE `shared_task_id` = {}", id);
		database.QueryDatabase(query);

		fmt::format_to(out, "INSERT INTO `shared_task_members` (shared_task_id, character_id, character_name, "
				    "is_leader) VALUES ");
		bool first = true;
		for (auto &&m : members.list) {
			if (first) {
				fmt::format_to(out, "({}, {}, \"{}\", {:d})", id, m.char_id, m.name, m.leader);
				first = false;
			} else {
				fmt::format_to(out, ", ({}, {}, \"{}\", {:d})", id, m.char_id, m.name, m.leader);
			}
		}
		query = fmt::to_string(out);
		out.clear();
		auto res = database.QueryDatabase(query);
		if (!res.Success())
			Log(Logs::General, Logs::Error, ERR_MYSQLERROR, res.ErrorMessage().c_str());
		else
			members.update = false;
	}

	database.TransactionCommit();

	// TODO: zone does some shit about completed tasks, is this for task history? Can we make zone do this?
}

/*
 * sets up activity stuff
 */
void SharedTask::InitActivities()
{
	task_state.TaskID = task_id;
	task_state.AcceptedTime = time(nullptr);
	task_state.Updated = true;
	task_state.CurrentStep = -1;

	for (int i = 0; i < shared_tasks.GetTaskActivityCount(task_id); i++) {
		task_state.Activity[i].ActivityID = i;
		task_state.Activity[i].DoneCount = 0;
		task_state.Activity[i].State = ActivityHidden;
		task_state.Activity[i].Updated = true;
	}
}

bool SharedTask::UnlockActivities()
{
	return true;
}

/*
 * Returns true if the task is completed
 */
bool SharedTask::TaskCompleted()
{
	auto task = shared_tasks.GetTaskInformation(task_id);

	if (task == nullptr)
		return false;

	for (int i = 0; i < task->ActivityCount; ++i) {
		if (!task->Activity[i].Optional && task_state.Activity[i].State != ActivityCompleted)
			return false;
	}

	return true;
}

/*
 * We just need to verify the activity can be updated, if it can, we tell zones
 * to do so. Update just means ticking up a count.
 *
 * We can safely throw away updates that would put us over the count
 *
 * zone has verified a lot of stuff, we're just doing it here to verify sync and shit
 */
void SharedTask::ProcessActivityUpdate(int activity_id, int value)
{
	auto task_info = shared_tasks.GetTaskInformation(task_id);
	// not shared task?
	if (task_info == nullptr)
		return;

	// OOB, throw it away
	if (activity_id > task_info->ActivityCount)
		return;

	// we're already done!
	if (task_state.Activity[activity_id].DoneCount == task_info->Activity[activity_id].GoalCount)
		return;

	task_state.Activity[activity_id].DoneCount =
	    std::min(task_state.Activity[activity_id].DoneCount + value, task_info->Activity[activity_id].GoalCount);
	if (task_state.Activity[activity_id].DoneCount >= task_info->Activity[activity_id].GoalCount)
		task_state.Activity[activity_id].State = ActivityCompleted;
	task_state.Activity[activity_id].Updated = true;

	// we just fire off to all zones, fuck it!
	auto pack = new ServerPacket(ServerOP_TaskActivityUpdate, sizeof(ServerSharedTaskActivity_Struct));
	auto update = (ServerSharedTaskActivity_Struct *)pack->pBuffer;
	update->id = id;
	update->activity_id = activity_id;
	update->value = task_state.Activity[activity_id].DoneCount;
	zoneserver_list.SendPacket(pack);
	safe_delete(pack);

	Save();

	if (TaskCompleted()) {
		SetCompleted(true);
		SetUpdated(true);
		auto pack = new ServerPacket(ServerOP_TaskCompleted, sizeof(uint32));
		pack->WriteUInt32(id);
		zoneserver_list.SendPacket(pack);
		safe_delete(pack);
		Save(); // since Save only saves stuff where update flag has been set, this isn't too bad to call again
	}

}

