<template>
	<transition name="back-to-top-fade">
		<div class="vue-back-to-top" v-show="visible" @click="backToTop" :title="text" :style="{bottom: bottom + 'px'}">
			<i class="fa fa-arrow-up"></i><span>&nbsp;&nbsp;{{text}}</span>
		</div>
	</transition>
</template>

<script>
export default {
	name: 'BackToTop',
	props: {
		text: {
			type: String,
			default: '返回顶部',
		},
		visibleOffset: {
			type: [String, Number],
			default: 600,
		},
		bottom: {
			type: Number,
			default: 45,
		},
	},
	data () {
		return {
			visible: false,
		}
	},
	/**
	 * Catch window scroll event
	 * @return {void}
	 */
	created () {
		window.smoothscroll = () => {
			let currentScroll = document.documentElement.scrollTop || document.body.scrollTop;
			if (currentScroll >= 5) {
				window.requestAnimationFrame(window.smoothscroll);
				window.scrollTo(0, currentScroll - (currentScroll / 5));
			} else if(currentScroll > 1) {
				window.requestAnimationFrame(window.smoothscroll);
				window.scrollTo(0, currentScroll - 1);
			} else if(currentScroll > 0) {
				window.scrollTo(0, 0);
			}
		}

		window.addEventListener("scroll", this.catchScroll);
	},
	beforeDestroy() {
		window.removeEventListener("scroll", this.catchScroll);
	},
	methods: {
		/**
		 * The function who make the magics
		 * @return {void}
		 */
		backToTop () {
			window.smoothscroll();
		},
		catchScroll() {
			this.visible = window.pageYOffset > parseInt(this.visibleOffset);
		},
	}
}
</script>

<style lang="less" scoped>
.back-to-top-fade-enter-active,
.back-to-top-fade-leave-active {
	transition: opacity .7s;
}
.back-to-top-fade-enter,
.back-to-top-fade-leave-to {
	opacity: 0;
}

.vue-back-to-top {
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

.sidebar-collapse .vue-back-to-top {
	left: 5px;
	width: 40px;

	span {
		display: none;
	}
}
</style>