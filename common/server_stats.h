#pragma once

#include <chrono>
#include <string>
#include <mutex>
#include <vector>
#include <uv.h>

namespace EQ
{
	struct ResourceUsageStats
	{
		double user_cpu_time;
		double system_cpu_time;
	};

	struct OSInfoStats
	{
		std::string system_name;
		std::string release;
		std::string version;
		std::string machine;
	};

	class ServerStats
	{
		typedef std::chrono::high_resolution_clock clock;
		typedef std::chrono::high_resolution_clock::time_point timestamp;
		typedef std::chrono::duration<double> seconds;
	public:
		~ServerStats() {
		}
		
		static ServerStats& Get() {
			static ServerStats inst;
			return inst;
		}
		
		void BeginFrame()
		{
			auto now = clock::now();
			auto inactive = now - m_last_end;
			m_inactive_time = std::chrono::duration_cast<seconds>(inactive).count();
			m_last_begin = now;
		}

		void EndFrame() {
			auto now = clock::now();
			auto active = now - m_last_begin;
			m_active_time = std::chrono::duration_cast<seconds>(active).count();
			m_last_end = now;
		}
		
		double InactiveFrameTime() const {
			return m_inactive_time;
		}

		double ActiveFrameTime() const {
			return m_active_time;
		}

		double FrameTime() const {
			return m_inactive_time + m_active_time;
		}

		void BeginNetworkFrame()
		{
			std::lock_guard<std::mutex> lock(m_network_lock);
			auto now = clock::now();
			auto inactive = now - m_network_last_end;
			m_network_inactive_time = std::chrono::duration_cast<seconds>(inactive).count();
			m_network_last_begin = now;
		}

		void EndNetworkFrame() 
		{
			std::lock_guard<std::mutex> lock(m_network_lock);
			auto now = clock::now();
			auto active = now - m_network_last_begin;
			m_network_active_time = std::chrono::duration_cast<seconds>(active).count();
			m_network_last_end = now;
		}

		double NetworkInactiveFrameTime() {
			std::lock_guard<std::mutex> lock(m_network_lock);
			return m_network_inactive_time;
		}

		double NetworkActiveFrameTime() {
			std::lock_guard<std::mutex> lock(m_network_lock);
			return m_network_active_time;
		}

		double NetworkFrameTime() {
			std::lock_guard<std::mutex> lock(m_network_lock);
			return m_network_inactive_time + m_network_active_time;
		}

		//OS Related Stuff
		int PID() {
			return uv_os_getpid();
		}

		int ParentPID() {
			return uv_os_getppid();
		}

		ResourceUsageStats ResourceUsage() {
			uv_rusage_t usage;
			if(0 != uv_getrusage(&usage)) {
				return ResourceUsageStats();
			}

			ResourceUsageStats ret;
			ret.user_cpu_time = static_cast<double>(usage.ru_utime.tv_sec) + static_cast<double>(usage.ru_utime.tv_usec / 1000000.0);
			ret.system_cpu_time = static_cast<double>(usage.ru_stime.tv_sec) + static_cast<double>(usage.ru_stime.tv_usec / 1000000.0);
			return ret;
		}

		OSInfoStats OSInfo() {
			uv_utsname_t osn;

			if (0 != uv_os_uname(&osn)) {
				return OSInfoStats();
			}

			OSInfoStats ret;
			ret.system_name = osn.sysname;
			ret.release = osn.release;
			ret.version = osn.version;
			ret.machine = osn.machine;
			return ret;
		}

		//Compile related things
		bool IsDebug() {
#ifdef NDEBUG
			return false;
#else
			return true;
#endif
		}
	private:
		ServerStats() {
			m_last_begin = clock::now();
			m_last_end = clock::now();
			m_network_last_begin = clock::now();
			m_network_last_end = clock::now();
		}

		//Main thread
		timestamp m_last_begin;
		timestamp m_last_end;
		double m_inactive_time;
		double m_active_time;

		//Network thread
		std::mutex m_network_lock;
		timestamp m_network_last_begin;
		timestamp m_network_last_end;
		double m_network_inactive_time;
		double m_network_active_time;
	};
}
