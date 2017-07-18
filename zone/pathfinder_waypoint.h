#pragma once

#include "pathfinder_interface.h"

class PathfinderWaypoint : public IPathfinder
{
public:
	PathfinderWaypoint() { }
	virtual ~PathfinderWaypoint() { }

	virtual IPath FindRoute(const glm::vec3 &start, const glm::vec3 &end);
	virtual glm::vec3 GetRandomLocation();
};