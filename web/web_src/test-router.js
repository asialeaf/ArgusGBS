import Vue from 'vue'
import Router from 'vue-router'
import store from './test-store'
import url from 'url'

import AdminLTE from 'components/AdminLTE.vue'

const TestRoot = () => import( /* webpackChunkName: 'test-root' */ 'components/TestRoot.vue')
const TestCatalog = () => import( /* webpackChunkName: 'test-catalog' */ 'components/TestCatalog.vue')
const TestRegister = () => import( /* webpackChunkName: 'test-register' */ 'components/TestRegister.vue')
const TestPlay = () => import( /* webpackChunkName: 'test-play' */ 'components/TestPlay.vue')
const TestPTZ = () => import( /* webpackChunkName: 'test-ptz' */ 'components/TestPTZ.vue')
const TestPreset = () => import( /* webpackChunkName: 'test-preset' */ 'components/TestPreset.vue')
const TestHomePosition = () => import( /* webpackChunkName: 'test-home-position' */ 'components/TestHomePosition.vue')
const TestPTZPrecise = () => import( /* webpackChunkName: 'test-ptz-precise' */ 'components/TestPTZPrecise.vue')
const TestSDCard = () => import( /* webpackChunkName: 'test-sdcard' */ 'components/TestSDCard.vue')
const TestTargetTrack = () => import( /* webpackChunkName: 'test-target-track' */ 'components/TestTargetTrack.vue')
const TestIFrame = () => import( /* webpackChunkName: 'test-iframe' */ 'components/TestIFrame.vue')
const TestPlayback = () => import( /* webpackChunkName: 'test-playback' */ 'components/TestPlayback.vue')
const TestRecord = () => import( /* webpackChunkName: 'test-record' */ 'components/TestRecord.vue')
const TestGuard = () => import( /* webpackChunkName: 'test-guard' */ 'components/TestGuard.vue')
const TestAlarmReset = () => import( /* webpackChunkName: 'test-alarm-reset' */ 'components/TestAlarmReset.vue')
const TestInfoStatus = () => import( /* webpackChunkName: 'test-info-status' */ 'components/TestInfoStatus.vue')
const TestTeleboot = () => import( /* webpackChunkName: 'test-teleboot' */ 'components/TestTeleboot.vue')
const TestConfig = () => import( /* webpackChunkName: 'test-config' */ 'components/TestConfig.vue')
const TestCruiseTrack = () => import( /* webpackChunkName: 'test-cruise-track' */ 'components/TestCruiseTrack.vue')
const TestSubscribe = () => import( /* webpackChunkName: 'test-subscribe' */ 'components/TestSubscribe.vue')
const TestBroadcast = () => import( /* webpackChunkName: 'test-broadcast' */ 'components/TestBroadcast.vue')
const TestUpgrade = () => import( /* webpackChunkName: 'test-upgrade' */ 'components/TestUpgrade.vue')

Vue.use(Router);

const router = new Router({
	routes: [
		{
			path: '/',
			component: AdminLTE,
			children: [{
				path: '',
				component: TestRoot,
				props: true,
				children: [{
					path: '',
					redirect: 'catalog'
				}, {
					path: 'catalog',
					component: TestCatalog,
					props: true,
				}, {
					path: 'register',
					component: TestRegister,
					props: true,
				}, {
					path: 'play',
					component: TestPlay,
					props: true,
				}, {
					path: 'ptz',
					component: TestPTZ,
					props: true,
				}, {
					path: 'preset',
					component: TestPreset,
					props: true,
				}, {
					path: 'homeposition',
					component: TestHomePosition,
					props: true,
				}, {
					path: 'ptzprecise',
					component: TestPTZPrecise,
					props: true,
				}, {
					path: 'sdcard',
					component: TestSDCard,
					props: true,
				}, {
					path: 'targettrack',
					component: TestTargetTrack,
					props: true,
				}, {
					path: 'iframe',
					component: TestIFrame,
					props: true,
				}, {
					path: 'playback',
					component: TestPlayback,
					props: true,
				}, {
					path: 'record',
					component: TestRecord,
					props: true,
				}, {
					path: 'guard',
					component: TestGuard,
					props: true,
				}, {
					path: 'alarmreset',
					component: TestAlarmReset,
					props: true,
				}, {
					path: 'infostatus',
					component: TestInfoStatus,
					props: true,
				}, {
					path: 'teleboot',
					component: TestTeleboot,
					props: true,
				}, {
					path: 'config',
					component: TestConfig,
					props: true,
				}, {
					path: 'cruisetrack',
					component: TestCruiseTrack,
					props: true,
				}, {
					path: 'subscribe',
					component: TestSubscribe,
					props: true,
				}, {
					path: 'broadcast',
					component: TestBroadcast,
					props: true,
				}, {
					path: 'upgrade',
					component: TestUpgrade,
					props: true,
				}]
			}, {
				path: 'logout',
				async beforeEnter(to, from, next) {
				await store.dispatch("logout");
					window.location.href = `/login`;
				}
			}, {
				path: 'blank',
				component: TestRoot,
			}, {
				path: '*',
				redirect: '/'
			}]
		}
	],
	linkActiveClass: 'active'
})

router.beforeEach(async (to, from, next) => {
	// let t1 = new Date();
	var serverInfo = await store.dispatch("getServerInfo");
	if(serverInfo) {
		document.title = serverInfo.LogoText || "LiveGBS";
	}
	// let t2 = new Date();
	// console.log("serverInfo cost %dms", t2.getTime() - t1.getTime());
	var userInfo = await store.dispatch("getUserInfo");
	// if (serverInfo && serverInfo.APIAuth === false && !userInfo) {
	//   next();
	//   return;
	// }
	// let t3 = new Date();
	// console.log("userInfo cost %dms", t3.getTime() - t2.getTime());
	var menuMap = store.state.menus.reduce((pval, cval) => {
		pval[cval.path] = cval;
		return pval;
	}, {})
	var pageRoles = []; // 前往页面要求角色列表
	if (to.meta.roles) {
		pageRoles.push(...to.meta.roles);
	}
	var menu = menuMap[to.path];
	if (menu && menu.roles) {
		pageRoles.push(...menu.roles);
	}
	if (menu && menu.title) {
		store.commit("updateTopic", menu.title);
	}
	if (!userInfo) {
		if ((serverInfo && serverInfo.APIAuth !== false) || to.matched.some(record => record.meta.needLogin)) {
			if (to.fullPath == '/') {
				window.location.href = `/login`;
			} else {
				var _url = url.parse(window.location.href, true);
				_url.hash = to.fullPath;
				window.location.href = `/login?r=${encodeURIComponent(url.format(_url))}`;
			}
			return;
		}
		next();
		return;
	}
	if (pageRoles.length == 0) {
		next();
		return;
	}
	if (!pageRoles.some(pr => (userInfo.Roles.some(ur => (ur == pr || ur == '超级管理员'))))) { // 两个角色列表没有交集
		console.log("page", to.path, "require roles", pageRoles.join(','));
		console.log("user", userInfo.Name, "has roles", userInfo.Roles.join(','));
		for(var menu of store.state.menus) { // 取首个有交集的菜单
			if(menu.path) {
				if(!menu.roles || menu.roles.some(pr => (userInfo.Roles.some(ur => (ur == pr))))) {
					next(menu.path);
					return;
				}
			}
		}
		next('/blank');
		return;
	}
	next();
})

export default router;