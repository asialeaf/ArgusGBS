<template>
	<transition name="sidebar-toggle-fade">
		<a class="sidebar-toggle" data-toggle="push-menu" role="button" v-show="visible" :title="text" :style="{bottom: bottom + 'px'}">
			<i class="fa fa-bars"></i><span>&nbsp;&nbsp;{{text}}</span>
		</a>
	</transition>
</template>

<script>
export default {
	name: "SidebarToggle",
	props: {
		text: {
			type: String,
			default: '折叠边栏',
		},
		visibleOffset: {
			type: [String, Number],
			default: 600,
		},
		bottom: {
			type: Number,
			default: 10,
		},
	},
	data() {
		return {
			visible: false,
		};
	},
	created() {
		window.addEventListener("scroll", this.catchScroll);
	},
	beforeDestroy() {
		window.removeEventListener("scroll", this.catchScroll);
	},
	methods: {
		catchScroll() {
			this.visible = window.pageYOffset > parseInt(this.visibleOffset);
		},
	},
};
</script>

<style lang="less" scoped>
.sidebar-toggle-fade-enter-active,
.sidebar-toggle-fade-leave-active {
	transition: opacity 0.7s;
}
.sidebar-toggle-fade-enter,
.sidebar-toggle-fade-leave-to {
	opacity: 0;
}

.sidebar-toggle {
	position: fixed;
	left: 0;
	width: 160px;
	color: #b8c7ce;
	text-align: center;
	line-height: 30px;
	background-color: transparent;
	border-radius: 3px;
	z-index: 1000;
	cursor: pointer;

	&:hover {
		color: #fff;
	}
}

.sidebar-collapse .sidebar-toggle {
	left: 5px;
	width: 40px;

	span {
		display: none;
	}
}
</style>