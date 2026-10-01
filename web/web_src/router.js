import Vue from 'vue'
import Router from 'vue-router'
import store from './store'
import url from 'url'

import AdminLTE from 'components/AdminLTE.vue'

const Dashboard = () => import( /* webpackChunkName: 'dashboard' */ 'components/Dashboard.vue')

const ContentRoot = () => import( /* webpackChunkName: 'root' */ 'components/ContentRoot.vue')

const Screen  = () => import(/* webpackChunkName: 'screen' */ 'components/Screen.vue')

const MapView = () => import(/* webpackChunkName: 'mapview' */ 'components/MapView.vue')

const DeviceList = () => import( /* webpackChunkName: 'device' */ 'components/DeviceList.vue')
const DeviceTree = () => import( /* webpackChunkName: 'device' */ 'components/DeviceTree.vue')
const ChannelList = () => import( /* webpackChunkName: 'device' */ 'components/ChannelList.vue')
const PlaybackListBox = () => import(/* webpackChunkName: 'device' */ 'components/PlaybackListBox.vue')
const PlaybackTimebox = () => import(/* webpackChunkName: 'device' */ 'components/PlaybackTimebox.vue')
const AlarmList = () => import(/* webpackChunkName: 'alarm' */ 'components/AlarmList.vue')
const AlarmPlanList = () => import(/* webpackChunkName: 'alarm' */ 'components/AlarmPlanList.vue')
const LogList = () => import(/* webpackChunkName: 'log' */ 'components/LogList.vue')
const CascadeList = () => import(/* webpackChunkName: 'cascade' */ 'components/CascadeList.vue')
const UserList = () => import(/* webpackChunkName: 'user' */ 'components/UserList.vue')
const Config = () => import(/* webpackChunkName: 'config' */ 'components/Config.vue')
const About = () => import(/* webpackChunkName: 'about' */ 'components/About.vue')

const CloudRecordChannelList = () => import(/* webpackChunkName: 'cloudrecord' */ 'components/CloudRecordChannelList.vue')
const CloudRecordListBox = () => import(/* webpackChunkName: 'cloudrecord' */ 'components/CloudRecordListBox.vue')
const CloudRecordTimeBox = () => import(/* webpackChunkName: 'cloudrecord' */ 'components/CloudRecordTimeBox.vue')
const CloudRecordPlanList = () => import(/* webpackChunkName: 'cloudrecord' */ 'components/CloudRecordPlanList.vue')

const WhiteList = () => import(/* webpackChunkName: 'white' */ 'components/WhiteList.vue')
const BlackList = () => import(/* webpackChunkName: 'black' */ 'components/BlackList.vue')
const GmReqList = () => import(/* webpackChunkName: 'gmreq' */ 'components/GmReqList.vue')


Vue.use(Router);

const router = new Router({
	routes: [
		{
			path: '/',
			component: AdminLTE,
			children: [{
				path: '',
				component: Dashboard,
				props: true
			}, {
				path: 'map',
				component: MapView,
				props: true
			}, {
				path: 'screen',
				component: ContentRoot,
				children: [
					{
						path: '',
						component: Screen,
						props: true
					}
				]
			}, {
				path: 'devices',
				component: ContentRoot,
				children: [
					{
						path: '',
						redirect: '1'
					}, {
						path: 'tree',
						component: DeviceTree,
						props: true,
					}, {
						path: ':page',
						component: DeviceList,
						props: true
					}, {
						path: 'channels/:devid/:page',
						component: ChannelList,
						props: true
					}, {
						path: 'playback/list/:devid/:channel/:day?',
						component: PlaybackListBox,
						props: true
					}, {
						path: 'playback/timebox/:devid/:channel/:day?',
						component: PlaybackTimebox,
						props: true
					}
				]
			}, {
				path: 'cloudrecord',
				component: ContentRoot,
				children: [
					{
						path: '',
						redirect: '1'
					}, {
						path: ':page',
						component: CloudRecordChannelList,
						props: true
					}, {
						path: 'listview/:serial/:code/:day?',
						component: CloudRecordListBox,
						props: true
					}, {
						path: 'timeview/:serial/:code/:day?',
						component: CloudRecordTimeBox,
						props: true
					}, {
						path: 'plans/:page',
						meta: { needLogin: true, roles: ['超级管理员'] },
						component: CloudRecordPlanList,
						props: true
					}
				]
			}, {
				path: 'alarms',
				component: ContentRoot,
				children: [
					{
						path: '',
						redirect: '1'
					}, {
						path: ':page',
						component: AlarmList,
						meta: { roles: ['管理员','操作员'] },
						props: true
					}, {
						path: 'plans/:page',
						meta: { needLogin: true, roles: ['超级管理员'] },
						component: AlarmPlanList,
						props: true
					}
				]
			}, {
				path: 'cascade',
				meta: { needLogin: true },
				component: ContentRoot,
				children: [
					{
						path: '',
						redirect: '1'
					}, {
						path: ':page',
						meta: { needLogin: true, roles: ['管理员'] },
						component: CascadeList,
						props: true
					}
				]
			}, {
				path: 'user',
				meta: { needLogin: true },
				component: ContentRoot,
				children: [
					{
						path: '',
						redirect: '1'
					}, {
						path: ':page',
						meta: { needLogin: true, roles: ['管理员'] },
						component: UserList,
						props: true
					}
				]
			}, {
				path: 'white',
				meta: { needLogin: true },
				component: ContentRoot,
				children: [
					{
						path: '',
						redirect: '1'
					}, {
						path: ':page',
						meta: { needLogin: true, roles: ['超级管理员'] },
						component: WhiteList,
						props: true
					}
				]
			}, {
				path: 'black',
				meta: { needLogin: true },
				component: ContentRoot,
				children: [
					{
						path: '',
						redirect: '1'
					}, {
						path: ':page',
						meta: { needLogin: true, roles: ['超级管理员'] },
						component: BlackList,
						props: true
					}
				]
			}, {
				path : 'logs',
				meta: { needLogin: true },
				component: ContentRoot,
				children: [
					{
						path: '',
						redirect: '1'
					}, {
						path: ':page',
						meta: { needLogin: true, roles: ['超级管理员'] },
						component: LogList,
						props: true
					}
				]
			}, {
				path: 'gmreq',
				redirect: 'gm/req'
			}, {
				path: 'gm/req',
				meta: { needLogin: true },
				component: ContentRoot,
				children: [
					{
						path: '',
						redirect: '1'
					}, {
						path: ':page',
						meta: { needLogin: true, roles: ['超级管理员'] },
						component: GmReqList,
						props: true
					}
				]
			}, {
				path: 'config',
				meta: { needLogin: true },
				component: ContentRoot,
				children: [
					{
						path: '',
						component: Config,
					}
				]
			}, {
				path: 'about',
				meta: { needLogin: true },
				component: ContentRoot,
				children: [
					{
						path: '',
						component: About,
					}
				]
			}, {
				path: 'logout',
				async beforeEnter(to, from, next) {
				await store.dispatch("logout");
					window.location.href = `/login`;
				}
			}, {
				path: 'blank',
				component: ContentRoot,
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