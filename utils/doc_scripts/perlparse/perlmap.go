package main

import ()

//perlReturnTypes are mapped to identify what sort of return the script does
var perlReturnTypes = map[string]string{
	"boolSV(":   "bool",
	"PUSHu(":    "uint",
	"PUSHi(":    "int",
	"sv_setpv(": "string",
	"PUSHn(":    "double",
}

//Paths are where every perl file is at
var perlPaths = []*path{
	{
		Name:    "../../../zone/embparser_api.cpp",
		Scope:   "General",
		Replace: "quest",
	},
	{
		Name:  "../../../zone/perl_client.cpp",
		Scope: "Client",
	},
	{
		Name:    "../../../zone/perl_doors.cpp",
		Scope:   "Doors",
		Replace: "door",
	},
	{
		Name:  "../../../zone/perl_entity.cpp",
		Scope: "EntityList",
	},
	{
		Name:  "../../../zone/perl_groups.cpp",
		Scope: "Group",
	},
	{
		Name:  "../../../zone/perl_hateentry.cpp",
		Scope: "HateEntry",
	},
	{
		Name:  "../../../zone/perl_mob.cpp",
		Scope: "Mob",
	},
	{
		Name:  "../../../zone/perl_npc.cpp",
		Scope: "NPC",
	},
	{
		Name:  "../../../zone/perl_object.cpp",
		Scope: "Object",
	},
	{
		Name:    "../../../zone/perl_perlpacket.cpp",
		Scope:   "PerlPacket",
		Replace: "packet",
	},
	{
		Name:  "../../../zone/perl_player_corpse.cpp",
		Scope: "Corpse",
	},
	{
		Name:  "../../../zone/perl_QuestItem.cpp",
		Scope: "QuestItem",
	},
	{
		Name:  "../../../zone/perl_raids.cpp",
		Scope: "Raid",
	},
	{
		Name:  "../../../zone/embparser.cpp",
		Scope: "Event",
	},
}

//These are known parameter types
var perlKnownTypes = map[string]string{
	"activity_id":               "uint",
	"alt_mode":                  "bool",
	"anim_num":                  "int",
	"best_z":                    "float",
	"buttons":                   "int",
	"channel_id":                "int",
	"char_id":                   "int",
	"charges":                   "int",
	"class_id":                  "int",
	"client_name":               "string",
	"color":                     "int",
	"color_id":                  "int",
	"condition_id":              "int",
	"copper":                    "int",
	"count":                     "int",
	"debug_level":               "int",
	"decay_time":                "int",
	"dest_heading":              "float",
	"dest_x":                    "float",
	"dest_y":                    "float",
	"dest_z":                    "float",
	"in_lastname":               "string",
	"distance":                  "int",
	"door_id":                   "int",
	"value":                     "int",
	"cost":                      "int",
	"slot":                      "int",
	"type":                      "int",
	"iSendToSelf":               "int",
	"iFromDB":                   "bool",
	"duration":                  "int",
	"effect_id":                 "int",
	"elite_material_id":         "int",
	"enforce_level_requirement": "bool",
	"explore_id":                "uint",
	"faction_value":             "int",
	"fade_in":                   "int",
	"fade_out":                  "int",
	"fadeout":                   "uint",
	"firstname":                 "string",
	"format":                    "string",
	"from":                      "string",
	"gender_id":                 "int",
	"gold":                      "int",
	"grid_id":                   "int",
	"guild_rank_id":             "int",
	"heading":                   "float",
	"hero_forge_model_id":       "int",
	"ignore_quest_update":       "bool",
	"instance_id":               "int",
	"int_unused":                "int",
	"int_value":                 "int",
	"is_enabled":                "bool",
	"is_strict":                 "bool",
	"item_id":                   "int",
	"key":                       "string",
	"language_id":               "int",
	"lastname":                  "string",
	"leader_name":               "string",
	"level":                     "int",
	"link_name":                 "string",
	"macro_id":                  "int",
	"max_level":                 "int",
	"max_x":                     "float",
	"max_y":                     "float",
	"max_z":                     "float",
	"message":                   "string",
	"milliseconds":              "int",
	"min_level":                 "int",
	"min_x":                     "float",
	"min_y":                     "float",
	"min_z":                     "float",
	"name":                      "string",
	"new_hour":                  "int",
	"new_min":                   "int",
	"node1":                     "int",
	"node2":                     "int",
	"npc_id":                    "int",
	"npc_type_id":               "int",
	"object_type":               "int",
	"options":                   "int",
	"platinum":                  "int",
	"popup_id":                  "int",
	"priority":                  "int",
	"quantity":                  "int",
	"race_id":                   "int",
	"remove_item":               "bool",
	"requested_id":              "int",
	"reset_base":                "bool",
	"saveguard":                 "bool",
	"seconds":                   "int",
	"send_to_world":             "bool",
	"signal_id":                 "int",
	"silent":                    "bool",
	"silver":                    "int",
	"size":                      "int",
	"spell_id":                  "int",
	"stat_id":                   "int",
	"str_value":                 "string",
	"subject":                   "string",
	"target_enum":               "string",
	"target_id":                 "int",
	"task":                      "int",
	"task_id":                   "uint",
	"task_id1":                  "int",
	"number":                    "int",
	"task_id10":                 "int",
	"task_id2":                  "int",
	"task_set":                  "int",
	"taskid":                    "int",
	"taskid1":                   "int",
	"taskid2":                   "int",
	"taskid3":                   "int",
	"taskid4":                   "int",
	"teleport":                  "int",
	"temp":                      "int",
	"texture_id":                "int",
	"theme_id":                  "int",
	"update_world":              "int",
	"updated_time_till_repop":   "uint",
	"version":                   "int",
	"wait_ms":                   "int",
	"window_title":              "string",
	"x":                         "float",
	"y":                         "float",
	"z":                         "float",
	"zone_id":                   "int",
	"zone_short":                "string",
	`task_id%i`:                 "int",
}

var perlKnownEventArguments = map[string]string{}

var perlKnownEventTypes = map[string]string{
	"activity_id":         "int",    //", sep.arg[1]);
	"caster_id":           "int",    //", extradata);
	"charid":              "int",    //", char_id);
	"class":               "int",    //", GetClassIDName(mob->GetClass()));
	"clicker_id":          "int",    //", extradata);
	"combat_state":        "int",    //", data);
	"copper":              "int",    //", GetVar("copper." + std::string(itoa(objid))).c_str());
	"corpse":              "int",    //", sep.arg[2]);
	"data":                "string", //", "0");
	"donecount":           "int",    //", sep.arg[0]);
	"doorid":              "int",    //", data);
	"env_damage":          "int",    //", sep.arg[0]);
	"env_damage_type":     "int",    //", sep.arg[1]);
	"env_final_damage":    "int",    //", sep.arg[2]);
	"faction":             "int",    //", itoa(fac));
	"fished_item":         "int",    //", extradata);
	"foraged_item":        "int",    //", extradata);
	"gold":                "int",    //", GetVar("gold." + std::string(itoa(objid))).c_str());
	"grouped":             "int",    //", mob->IsGrouped());
	"h":                   "int",    //", npcmob->GetHeading() );
	"hate_state":          "int",    //", data);
	"hpevent":             "int",    //", "-1");
	"hpratio":             "int",    //",npcmob->GetHPRatio());
	"inchpevent":          "int",    //", "-1");
	"instanceid":          "int",    //", zone->GetInstanceID());
	"instanceversion":     "int",    //", zone->GetInstanceVersion());
	"itemid":              "int",    //", extradata);
	"itemname":            "string", //", item_inst->GetItem()->Name);
	"killed":              "int",    //", mob->GetNPCTypeID());
	"killed_npc_id":       "int",    //", sep.arg[4]);
	"killer_damage":       "int",    //", sep.arg[1]);
	"killer_id":           "int",    //", sep.arg[0]);
	"killer_skill":        "int",    //", sep.arg[3]);
	"killer_spell":        "int",    //", sep.arg[2]);
	"langid":              "int",    //", "0");
	"looted_charges":      "int",    //", sep.arg[1]);
	"looted_id":           "int",    //", sep.arg[0]);
	"mlevel":              "int",    //", npcmob->GetLevel());
	"mname":               "string", //", npcmob->GetName());
	"mobid":               "int",    //", npcmob->GetID());
	"name":                "string", //", mob->GetName());
	"objectid":            "int",    //", data);
	"option":              "int",    //", data);
	"picked_up_entity_id": "int",    //", extradata);
	"picked_up_id":        "int",    //", data);
	"platinum":            "int",    //", GetVar("platinum." + std::string(itoa(objid))).c_str());
	"popupid":             "int",    //", data);
	"quantity":            "int",    //", item_inst->IsStackable() ? item_inst->GetCharges() : 1);
	"race":                "int",    //", GetRaceIDName(mob->GetRace()));
	"raided":              "int",    //", mob->IsRaidGrouped());
	"recipe_id":           "int",    //", extradata);
	"recipe_name":         "string", //", data);
	"resurrect":           "int",    //", extradata);
	"signal":              "int",    //", data);
	"silver":              "int",    //", GetVar("silver." + std::string(itoa(objid))).c_str());
	"skill_id":            "int",    //", sep.arg[0]);
	"skill_level":         "int",    //", sep.arg[1]);
	"slotid":              "int",    //", extradata);
	"spawned_entity_id":   "int",    //", sep.arg[0]);
	"spawned_npc_id":      "int",    //", sep.arg[1]);
	"spell_id":            "int",    //", data);
	"status":              "int",    //", mob->CastToClient()->Admin());
	"target_zone_id":      "int",    //", data);
	"targetid":            "int",    //", npcmob->GetTarget()->GetID());
	"targetname":          "string", //", npcmob->GetTarget()->GetName());
	"task_id":             "int",    //", data);
	"text":                "string", //", data);
	"timer":               "int",    //", data);
	"uguild_id":           "int",    //", mob->CastToClient()->GuildID());
	"uguildrank":          "int",    //", mob->CastToClient()->GuildRank());
	"ulevel":              "int",    //", mob->GetLevel());
	"userid":              "int",    //", mob->GetID());
	"version":             "int",    //", zone->GetInstanceVersion());
	"wp":                  "int",    //", data);
	"x":                   "int",    //", npcmob->GetX() );
	"y":                   "int",    //", npcmob->GetY() );
	"z":                   "int",    //", npcmob->GetZ() );
	"zonehour":            "int",    //", eqTime.hour - 1);
	"zoneid":              "int",    //", zone->GetZoneID());
	"zoneln":              "string", //", zone->GetLongName());
	"zonemin":             "int",    //", eqTime.minute);
	"zonesn":              "string", //", zone->GetShortName());
	"zonetime":            "int",    //", (eqTime.hour - 1) * 100 + eqTime.minute);
	"zoneweather":         "int",    //", zone->zone_weather);
}
