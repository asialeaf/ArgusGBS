<template>
	<aside class="main-sidebar">
		<section class="sidebar">
			<ul class="sidebar-menu">
				<template v-for="(item,index) in menus">
					<router-link class="treeview" :key="index" :to="item.path" tag="li" :exact="item.path == '/'" v-if="showMenu(item)">
						<a>
							<i :class="['fa', 'fa-' + item.icon]"></i>
							<span>{{item.title}}</span>
						</a>
					</router-link>
				</template>
			</ul>
		</section>
	</aside>
</template>

<script>
export default {
	props: {
		menus : {
            default : () => []
		},
        serverInfo: {
            type: Object,
            default: () => {}
        },
        userInfo: {
            type: Object,
            default: () => null
        }
	},
	computed: {
		path() {
			return location.pathname;
		}
	},
	methods: {
		showMenu(item) {
			if(item.versionType && item.versionType != this.serverInfo.VersionType) {
				return false;
			}
			if(item.notVersionType && item.notVersionType == this.serverInfo.VersionType) {
				return false;
			}
			if(item.roles && this.userInfo && !this.hasAnyRole(this.serverInfo, this.userInfo, ...item.roles)) {
				return false;
			}
			if(item.path == "/about" && this.serverInfo.ShowAbout === false) {
				return false;
			}
			if(item.path == "/map" && !this.serverInfo.MapInfo) {
				return false;
			}
			if(item.hideInIE && videojs.browser.IE_VERSION) {
				return false;
			}
			return true;
		}
	}
}
</script>

<style lang="less" scoped>
.main-sidebar {
    /* Fix for IE */
    -webkit-transition: none;
    -o-transition: none;
    transition: none;
}
</style>
