<template>
	<header class="main-header">
		<router-link class="logo" style="position: relative;" to="/">
            <span class="logo-mini">{{serverInfo.LogoMiniText || 'GBS'}}</span>
            <span class="logo-lg">{{serverInfo.LogoText || 'LiveGBS'}}</span>
            <!-- <span class='logo-lg label label-warning' style='position: absolute;top:8px;right: 20px;font-size: 8px;' v-if="versionText">{{versionText}}</span> -->
		</router-link>

		<nav class="navbar navbar-static-top">
			<a class="sidebar-toggle" data-toggle="push-menu" role="button">
				<span class="sr-only">Toggle navigation</span>
			</a>
			<ul class="nav navbar-nav hidden-xs" v-if="serverInfo.LogoLongText">
				<li class="logo-long">{{serverInfo.LogoLongText}}</li>
			</ul>
			<div class="navbar-custom-menu">
                <ul class="nav navbar-nav">
                    <li v-if="hasAnyRole(serverInfo, userInfo, '管理员') && serverInfo.StreamStatsInterval > 0 && streamTotal > 0" class="net-show hidden-xs">
                        <a @click.prevent="$emit('show-stream-stats')" role="button">
                            开流率 ：{{streamOnline}}/{{streamTotal}} <br>
                            百分比 ：{{streamOnline == 0 ? 0 : parseFloat((streamOnline/streamTotal*100).toFixed(2))}}%
                        </a>
                    </li>
                    <li v-if="hasAnyRole(serverInfo, userInfo, '超级管理员')" class="net-show hidden-xs">
                        <a @click.prevent="$emit('show-session-list')" role="button">
                            接收 <i class="glyphicon glyphicon-arrow-up"></i> ：{{parseFloat(recvRate).toFixed(3)}} Mbps <br>
                            发送 <i class="glyphicon glyphicon-arrow-down"></i> ：{{parseFloat(sentRate).toFixed(3)}} Mbps
                        </a>
                    </li>
                    <li v-if="serverInfo.DemoUser">
                        <a target="_blank" href="//www.liveqing.com/docs/download/LiveGBS.html"><i class="fa fa-download"></i> 下载使用</a>
                    </li>
                    <li v-if="serverInfo.DemoUser" class="hidden-xs">
                        <a target="_blank" href="/apidoc/"><i class="fa fa-book"></i> 开发接口</a>
                    </li>
                    <li class="dropdown" v-if="userInfo">
                        <a href="#" class="dropdown-toggle" data-toggle="dropdown">欢迎, {{userInfo.Name}}</a>
                        <ul class="dropdown-menu">
                            <li v-if="!serverInfo.DemoUser && hasAnyRole(serverInfo, userInfo, '超级管理员')">
                                <a target="_blank" href="/apidoc/"><i class="fa fa-book"></i> 开发接口</a>
                            </li>
                            <li v-if="!userInfo.Cas && !userInfo.OAuth">
                                <a @click.prevent="$emit('modify-password')" role="button"><i class="fa fa-key"></i> 修改密码</a>
                            </li>
                            <li>
                                <a href="/logout"><i class="fa fa-sign-out"></i> 注 销</a>
                            </li>
                            <!-- <router-link to="/logout" tag="li">
                                <a><i class="fa fa-sign-out"></i> 注 销</a>
                            </router-link> -->
                        </ul>
                    </li>
                    <li v-else>
                        <!-- <a href="/login.html"><i class="fa fa-user"></i> 登录</a> -->
                        <a href="/login"><i class="fa fa-user"></i> 登录</a>
                    </li>
                </ul>
			</div>
		</nav>
	</header>
</template>

<script>
export default {
	data() {
        return {
            timer: 0,
            sentRate: 0,
            recvRate: 0,
            bNetStating: false,
            streamTimer: 0,
            streamOnline: 0,
            streamTotal: 0,
            bStreamStating: false,
        };
	},
	props: {
        serverInfo: {
            type: Object,
            default: () => {}
        },
        userInfo: {
            type: Object,
            default: () => null
        }
	},
	mounted() {
		if (this.hasAnyRole(this.serverInfo, this.userInfo, '超级管理员')) {
			this.netStats();
			this.timer = setInterval(() => {
				this.netStats();
			}, 3000);
		}
        if (this.hasAnyRole(this.serverInfo, this.userInfo, '管理员') && this.serverInfo.StreamStatsInterval > 0) {
            this.streamStats();
            this.streamTimer = setInterval(() => {
                this.streamStats();
            }, this.serverInfo.StreamStatsInterval * 1000);
        }
	},
	beforeDestroy() {
        if (this.timer) {
            clearInterval(this.timer);
            this.timer = 0;
        }
        if (this.streamTimer) {
            clearInterval(this.streamTimer);
            this.streamTimer = 0;
        }
	},
	methods: {
		netStats() {
            if(this.bNetStating) return;
            this.bNetStating = true;
			$.ajax({
                method: "GET", // type: "GET", is also ok, method since jquery 1.9
                url: "/api/v1/dashboard/top/net",
                global: false,
			}).then(ret => {
                if (ret) {
                    this.sentRate = ret.sent;
                    this.recvRate = ret.recv;
                }
			}).fail(xhr => {
				if (xhr) {
					console.log(`stats net error: ${xhr.status} ${xhr.responseText}`);
					if (xhr.status == 401 && this.serverInfo.DemoUser) {
						location.href = `/login?r=${encodeURIComponent(window.location.href)}`;
						return false;
					}
				}
			}).always(() => {
                this.bNetStating = false;
            });
		},
        streamStats() {
            if(this.bStreamStating) return;
            this.bStreamStating = true;
            $.ajax({
                method: "GET",
                url: "/api/v1/device/streamstats",
                global: false,
            }).then(ret => {
                if (ret && ret.List && ret.List.length > 0) {
                    this.streamOnline = ret.List[0].Online;
                    this.streamTotal = ret.List[0].Total;
                }
            }).always(() => {
                this.bStreamStating = false;
            });
        }
	}
}
</script>

<style lang="less" scoped="true">
.main-header .navbar,
.main-header .logo {
	/* Fix for IE */
	-webkit-transition: none;
	-o-transition: none;
	transition: none;
}

.main-header .logo {
	padding: 0;
}

.main-header .logo-long {
    font-size: 18px;
    text-align: left;
    color: white;
    font-weight: 100;
    line-height: 18px;
    padding: 16px;
}

.main-header .navbar .dropdown-menu li a {
	color: #777;
	&:hover {
		background-color: #e1e3e9;
	}
}

.main-header .net-show {
    font-size: 12px;
    text-align: left;
    color: white;
    opacity: 0.5;
    font-weight: 100;
    line-height: 16px;

    // 9 + 9 + 16 + 16 = 50
    & > a {
        padding-top: 9px !important;
        padding-bottom: 9px !important;
        line-height: 16px !important;
    }
}
</style>
