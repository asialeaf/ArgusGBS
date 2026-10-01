<template>
<div :class="['modal', { fade: fade }]" data-keyboard="true" data-backdrop="static" tabindex="-1">
    <div :class="['modal-dialog', size]">
        <div class="modal-content">
            <div class="modal-header">
                <button type="button" class="close" data-dismiss="modal" aria-label="Close" title="关闭">
                    <span aria-hidden="true">&times;</span>
                </button>
                <a role="button" class="ptz" @click="togglePTZVisible" title="云台控制" v-show="showPTZButton">
                    <span aria-hidden="true">
                        <i class="fa fa-plus" v-if="!ptzVisible"></i>
                        <i class="fa fa-minus" v-else></i>
                    </span>
                </a>
                <h4 class="modal-title text-primary text-center"><span>{{videoTitle}}</span></h4>
            </div>
            <div class="modal-body">
                <div class="play-area" @mousemove="resetActiveTimer" @touchstart="resetActiveTimer">
                    <LivePlayer ref="player" v-if="bShow" :videoUrl="videoUrl" :waterMark="osd"
                        :poster="snapUrl" :hideBigPlayButton="hideBigPlayButton !== undefined ? hideBigPlayButton : !!serverInfo.HideBigPlayButton" live
                        :controls="controls" :showCameraButton="snap" :showCustomButton="custom" :alt="alt" :debug="debug"
                        :fluent="fluent" :stretch="stretch" :autoplay="autoplay" :muted="isMuted" :digitalZoom="digitalZoom" :dblclickFullscreen="dblclickFullscreen"
                        @media_info="onMediaInfo" @ended="onEnded" @error="onError" @message="$message" @fullscreen="onFullscreenChange"
                        :loading.sync="bLoading" v-loading="bLoading" element-loading-text="加载中..." element-loading-background="#000">
                        <div class="ptz-block-fs" v-show="showPTZBlock">
                            <div class="ptz-cell ptz-up" @mousedown.prevent="ptzControl('up', $event)" title="上">
                                <i class="fa fa-chevron-up"></i>
                            </div>
                            <div class="ptz-cell ptz-left" @mousedown.prevent="ptzControl('left', $event)" title="左">
                                <i class="fa fa-chevron-left"></i>
                            </div>
                            <div class="ptz-cell ptz-center" title="云台控制">
                                <i class="fa fa-arrows"></i>
                            </div>
                            <div class="ptz-cell ptz-right" @mousedown.prevent="ptzControl('right', $event)" title="右">
                                <i class="fa fa-chevron-right"></i>
                            </div>
                            <div class="ptz-cell ptz-down" @mousedown.prevent="ptzControl('down', $event)" title="下">
                                <i class="fa fa-chevron-down"></i>
                            </div>
                            <div class="ptz-cell ptz-zoomin" @mousedown.prevent="ptzControl('zoomin', $event)" title="放大">
                                <i class="fa fa-plus-circle"></i>
                            </div>
                            <div class="ptz-cell ptz-talk" @mousedown.prevent="talkStart" v-if="showTalk">
                                <i class="fa fa-microphone"></i>
                            </div>
                            <div class="ptz-cell ptz-zoomout" @mousedown.prevent="ptzControl('zoomout', $event)" title="缩小">
                                <i class="fa fa-minus-circle"></i>
                            </div>
                        </div>
                    </LivePlayer>
                </div>
                <div class="text-center" v-if="isDemoUser(serverInfo, userInfo) && !bOutHevcTip">
                    <br>
                    提示: 演示系统限制匿名登录播放时间, 若需测试长时间播放, 请<a target="_blank" href="//www.liveqing.com/docs/download/LiveGBS.html">下载使用</a>
                    <br>
                    <br>
                </div>
                <div class="text-center text-red" v-if="bOutHevcTip">
                    <br>
                    提示: 正在播放 H265 直出流, 确保浏览器版本较新, 并且开启硬件加速
                    <br>
                    <br>
                </div>
                <div class="ptz-block-bottom" v-show="showPTZBottom">
                    <el-progress class="mic-level-bar" :percentage="level" :text-inside="true" v-if="ws"></el-progress>
                    <div class="ptz-cell ptz-up" command="up" @mousedown.prevent="ptzControl('up', $event)" @touchstart.prevent="ptzControl('up', $event)" title="上">
                        <i class="fa fa-chevron-up"></i>
                    </div>
                    <div class="ptz-cell ptz-left" command="left" @mousedown.prevent="ptzControl('left', $event)" @touchstart.prevent="ptzControl('left', $event)" title="左">
                        <i class="fa fa-chevron-left"></i>
                    </div>
					<el-tooltip content="长按对讲" placement="top" manual :value="bTalkTip">
						<div class="ptz-cell ptz-center" title="云台控制">
							<i class="fa fa-microphone" title="按住喊话" @mousedown.prevent="talkStart" @touchstart.prevent="talkStart" v-if="showTalk"></i>
						</div>
					</el-tooltip>
                    <div class="ptz-cell ptz-right" command="right" @mousedown.prevent="ptzControl('right', $event)" @touchstart.prevent="ptzControl('right', $event)" title="右">
                        <i class="fa fa-chevron-right"></i>
                    </div>
                    <div class="ptz-cell ptz-down" command="down" @mousedown.prevent="ptzControl('down', $event)" @touchstart.prevent="ptzControl('down', $event)" title="下">
                        <i class="fa fa-chevron-down"></i>
                    </div>
                    <div class="ptz-cell ptz-zoomin" command="zoomin" @mousedown.prevent="ptzControl('zoomin', $event)" @touchstart.prevent="ptzControl('zoomin', $event)" title="放大">
                        <i class="fa fa-plus"></i>
                    </div>
                    <div class="ptz-cell ptz-zoomout" command="zoomout" @mousedown.prevent="ptzControl('zoomout', $event)" @touchstart.prevent="ptzControl('zoomout', $event)" title="缩小">
                        <i class="fa fa-minus"></i>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
</template>

<script>
import "jquery-ui/ui/widgets/draggable";
import LivePlayer from "@liveqing/liveplayer";

export default {
	data() {
		return {
			videoUrl: "",
			videoTitle: "",
			protocol_: "",
			osd: "",
			snapUrl: "",
			serial: "",
			code: "",
			streamid: "",
			streamInfo: null,
			bShow: false,
			bLoading: false,
			recorder: null,
			ws: null,
			level: 0,
			mediaInfo: null,
			bAudioEnable: false,
			bOutHevcTip: false,
			bTalkTip: false,
			bControls: false,
			alt: "无信号",
			fullscreen: false,
			ptzVisible: false,
			ptzType: 0,
			activeTimer: null,
			activeInterval: 0,
		};
	},
	props: {
		fade: {
			type: Boolean,
			default: false,
		},
		size: {
			type: String,
			default: '',
		},
		protocol: {
			type: String,
			default: '',
		},
		talk: {
			type: Boolean,
			default: false,
		},
		ptz: {
			type: Boolean,
			default: false,
		},
		token: {
			type: String,
			default: '',
		},
		controls: {
			type: Boolean,
			default: true,
		},
		snap: {
			type: Boolean,
			default: undefined,
		},
		custom: {
			type: Boolean,
			default: true,
		},
		fluent: {
			type: Boolean,
			default: true,
		},
		stretch: {
			type: Boolean,
			default: false,
		},
		autoplay: {
			type: Boolean,
			default: true,
		},
		muted: {
			type: String,
			default: '',
		},
		debug: {
			type: Boolean,
			default: false,
		},
		digitalZoom: {
			type: Boolean,
			default: true,
		},
		dblclickFullscreen: {
			type: Boolean,
			default: true,
		},
		hideBigPlayButton: {
			type: Boolean,
			default: undefined,
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
		showTalk() {
			return this.talk && this.canTalk() && this.serverInfo.VersionType != '标准版';
		},
		canPTZ() {
			return this.videoUrl && (this.ptzType === 0 || this.ptzType === 1);
		},
		showPTZButton() {
			return this.videoUrl && (this.ptz || this.talk) && this.hasAnyRole(this.serverInfo, this.userInfo, '管理员', '操作员');
		},
		showPTZBottom() {
			return this.ptzVisible && this.videoUrl;
		},
		showPTZBlock() {
			return this.ptzVisible && this.videoUrl && this.fullscreen && !this.isMobile() && this.bControls;
		},
		isMuted() {
			switch(this.muted) {
				case "yes":
					return true;
				case "no":
					return false;
				default:
					return !this.bAudioEnable;
			}
		}
	},
	beforeDestroy() {
		if(this.ws) {
			this.ws.close();
			this.ws = null;
		}
		if(this.activeTimer) {
			clearTimeout(this.activeTimer);
			this.activeTimer = null;
		}
		this.ctrlStop();
		$(document).off("mouseup touchend", this.mouseUp);
	},
	mounted() {
		$(this.$el).find(".modal-content").draggable({
			handle: ".modal-header",
			cancel: ".modal-title span, .close, .ptz",
			addClasses: false,
			containment: "document",
			delay: 100,
			opacity: 0.5
		});
		$(this.$el).on("hide.bs.modal", () => {
			this.ptzVisible = false;
			this.ptzType = 0;
			this.streamInfo = null;
			this.mediaInfo = null;
			this.snapUrl = "";
			this.videoUrl = "";
			this.videoTitle = "";
			this.protocol_ = "";
			this.osd = "";
			this.alt = "无信号";
			this.ctrlStop();
			this.bAudioEnable = false;
			this.bOutHevcTip = false;
			this.bTalkTip = false;
			this.bControls = false;
			this.bShow = false;
			if(this.activeTimer) {
				clearTimeout(this.activeTimer);
				this.activeTimer = null;
			}
		}).on("show.bs.modal", () => {
			this.bShow = true;
		});
		$(document).on("mouseup touchend", this.mouseUp);
	},
	components: { LivePlayer },
	methods: {
		play(serial, code, streamInfo) {
			this.serial = serial || "";
			this.code = code || "";
			streamInfo = streamInfo || {};
			var videoUrl = streamInfo.HLS;
			var snapUrl = streamInfo.SnapURL;
			var protocol = "HLS";
			if(this.flvSupported()) {
                videoUrl = this.isIE() ? streamInfo.WS_FLV : streamInfo.FLV;
                protocol = this.isIE() ? "WS_FLV" : "FLV";
			}
			switch(String(this.protocol || this.serverInfo.PreferStreamFmt).toUpperCase()) {
				case "WEBRTC":
					if(this.rtcSupported()) {
						videoUrl = streamInfo.WEBRTC;
						protocol = "WEBRTC";
					}
					break;
				case "WHEP":
					if(this.rtcSupported()) {
						videoUrl = streamInfo.WHEP;
						protocol = "WHEP";
					}
					break;
				case "FLV":
					if(this.flvSupported() && !this.isIE()) {
                        videoUrl = streamInfo.FLV;
                        protocol = "FLV";
					}
					break;
				case "WS_FLV":
				case "WS-FLV":
					if(this.flvSupported()) {
						videoUrl = streamInfo.WS_FLV;
						protocol = "WS_FLV";
					}
					break;
				case "HLS":
					videoUrl = streamInfo.HLS;
					protocol = "HLS";
					break;
				case "RTMP":
					videoUrl = streamInfo.RTMP;
					snapUrl = "";
					protocol = "RTMP";
					break;
			}
			this.streamid = streamInfo.StreamID || "";
			this.streamInfo = streamInfo;
			this.bAudioEnable = !!streamInfo.AudioEnable;
			this.videoTitle = streamInfo.ChannelName || code || "";
			this.osd = streamInfo.ChannelOSD || "";
			this.ptzType = streamInfo.ChannelPTZType || 0;
			this.protocol_ = protocol;
			if(this.token) {
				if(videoUrl) {
					if (videoUrl.indexOf("?") == -1) {
						videoUrl += `?token=${this.token}`;
					} else {
						videoUrl += `&token=${this.token}`;
					}
				}
				if(snapUrl) {
					if (snapUrl.indexOf("?") == -1) {
						snapUrl += `?token=${this.token}`;
					} else {
						snapUrl += `&token=${this.token}`;
					}
				}
			}
			this.snapUrl = snapUrl || "";
			this.videoUrl = videoUrl || ""; // no need in next tick since player@2.6.9
			if(!videoUrl && streamInfo.Message && typeof streamInfo.Message === "string") {
				this.alt = streamInfo.Message;
			}
			$(this.$el).modal("show");
		},
		ptzControl(cmd, event) {
			$.ajax({
				method: "POST",
				url: "/api/v1/control/ptz",
				global: false,
				data: {
					serial: this.serial,
					code: this.code,
					command: cmd,
					token: this.token,
				}
			}).fail(xhr => {
				if(xhr) {
					console.log(`ptz ${cmd} ajax error: ${xhr.status} ${xhr.responseText}`);
					if(xhr.status == 401 && this.$route && this.$route.path) {
						location.href = `/login?r=${encodeURIComponent(window.location.href)}`;
					}
				}
			})
			$(event.target).closest('.ptz-cell').addClass("active");
		},
		ptzStop() {
			if ($(this.$el).find(".ptz-cell.active").length > 0) {
				$.ajax({
					method: "POST",
					url: "/api/v1/control/ptz",
					global: false,
					data: {
						serial: this.serial,
						code: this.code,
						command: "stop",
						token: this.token,
					}
				}).fail(xhr => {
					if(xhr) {
						console.log(`ptz stop ajax error: ${xhr.status} ${xhr.responseText}`);
						if(xhr.status == 401 && this.$route && this.$route.path) {
							location.href = `/login?r=${encodeURIComponent(window.location.href)}`;
						}
					}
				}).always(() => {
					if(!this.isMobile() && this.fullscreen && this.bControls) {
						this.activeInterval = 5000;
						this.resetActiveTimer();
					}
				})
				$(this.$el).find(".ptz-cell.active").removeClass("active");
			}
		},
		talkStart(e) {
			if(!this.ws) {
				this.ws = new WebSocket(this.wsTalkURL());
				this.ws.onopen = evt => {
					console.log("ws talk open");
				}
				this.ws.onclose = evt => {
					console.log("ws talk close");
				}
				this.ws.onerror = evt => {
					console.log("ws talk error", evt);
				}
			} else if(this.serverInfo.TalkHold) {
				this.bTalkTip = false;
				this.talkStop();
				return;
			}
			var $target = $(e.currentTarget);
			if(this.recorder) {
				this.bTalkTip = false;
				$target.addClass("active");
				this.recorder.start();
				return;
			}
			LiveRecorder.get((rec, err) => {
				if(err) {
					alert(err);
					this.talkStop();
					return
				}
				// this.$refs["player"].setMuted(true);
				this.bTalkTip = (!this.ws && !this.serverInfo.TalkHold && !this.fullscreen);
				$target.addClass("active");
				this.recorder = rec;
				this.recorder.start();
			}, {
				sampleBits: 16,
				sampleRate: 8000,
				pcmCallback: (pcm, level) => {
					// binary to base64 string
					var reader = new window.FileReader();
					reader.onloadend = () => {
						var base64 = reader.result;
						var base64 = base64.split(',')[1];
						if(this.ws && this.ws.readyState === WebSocket.OPEN) {
							this.ws.send(base64);
							this.level = level;
						}
					}
					reader.readAsDataURL(pcm);
				}
			})
		},
		talkStop() {
			if(this.recorder) {
				// this.recorder.stop();
				this.recorder.destroy();
				this.recorder = null;
				this.bTalkTip = false;
				$(this.$el).find(".fa-microphone.active").removeClass("active");
				// this.$refs["player"].setMuted(false);
			}
			if(this.ws) {
				this.ws.close();
				this.ws = null;
				if(!this.isMobile() && this.fullscreen && this.bControls) {
					this.activeInterval = 5000;
					this.resetActiveTimer();
				}
			}
			this.level = 0;
		},
		ctrlStop() {
			this.talkStop();
			this.ptzStop();
		},
		mouseUp() {
			if(!this.serverInfo.TalkHold) {
				this.talkStop();
			}
			this.ptzStop();
		},
		wsTalkURL() {
			var protocol = "ws:";
			if(location.protocol.startsWith("https")) {
				protocol = "wss:";
			}
			//url query param "format=pcm|g711a"
			var ret = `${protocol}//${location.host}/api/v1/control/ws-talk/${this.serial}/${this.code}?format=pcm`;
			if(this.token) {
				ret += `&token=${this.token}`;
			}
			return ret;
		},
		onMediaInfo(mi) {
			this.mediaInfo = mi;
		},
		onEnded(e) {
			this.mediaInfo = null;
		},
		onError(e) {
			if(e == 'MediaError' && ((this.mediaInfo && String(this.mediaInfo.videoCodec).startsWith("hvc")) || this.protocol_ == "HLS")) {
				if(flvjs.getFeatureList() && !flvjs.getFeatureList().nativeMP4H265Playback) {
					this.bOutHevcTip = true;
					console.log("提示: 正在播放 H265 直出流, 确保浏览器版本较新, 并且开启硬件加速");
				}
			}
		},
		onFullscreenChange(fullscreen) {
			this.fullscreen = fullscreen;
			if(fullscreen) {
				$(this.$el).find(".ptz-block-fs").draggable({
					handle: '.ptz-center',
					// cancel: ".ptz-cell",
					containment: '.play-area .video-js',
					delay: 100
				});
			} else {
				$(this.$el).find(".ptz-block-fs").draggable('destroy');
			}
			this.$emit("fullscreen", fullscreen);
		},
		togglePTZVisible() {
			this.ptzVisible = !this.ptzVisible;
		},
		resetActiveTimer() {
			this.bControls = true;
			if (this.activeTimer) {
				clearTimeout(this.activeTimer)
			}
			this.activeTimer = setTimeout(() => {
				this.activeInterval = 0;
				if ($(this.$el).find(".ptz-block-fs .ptz-cell.active").length == 0) {
					this.bControls = false;
				}
			}, this.activeInterval || 2000);
		},
	}, //-- methods
};
</script>

<style lang="less" scoped>
.modal-dialog {
	margin-top: 110px;
	margin-bottom: 0;
}

.modal-content {
	border-radius: 6px;
}

.modal-title {
	overflow: hidden;
	white-space: nowrap;
	text-overflow: ellipsis;
}

.modal-header {
	padding: 4px;
	.ptz {
		position: absolute;
		top: 5px;
		right: 35px;
		margin-top: 0;
		line-height: 24px;
		color: #000;
		opacity: .2;

		&:hover, &:focus {
			color: #000;
			opacity: .5;
		}
	}
	.close {
		position: absolute;
		top: 5px;
		right: 10px;
		margin-top: 0;
		line-height: 24px;
	}
}
.modal-body {
	padding: 0;
}

.ptz-block-bottom {
	width: 150px;
	height: 210px;
	margin: 0 auto;
	margin-top: 10px;
	text-align: center;
	position: relative;
	font-size: 24px;

	.ptz-cell {
		width: 50px;
		height: 50px;
		line-height: 50px;
		position: absolute;
	}

	.ptz-cell.active {
		color: #ccc;
		font-size: 26px;
	}

	.fa-microphone, .fa-microphone-slash {
		padding: 10px 13px;
		border-radius: 25px;
	}

	.fa-microphone.active {
		color: #fff;
	}

	.mic-level-bar {
		position: absolute;
		transform: rotate(-90deg);
		width: 100%;
		top: 70px;
		left: -80px;
	}

	.ptz-center {
		top: 50px;
		left: 50px;
		border-radius: 25px;
		background-color: #ccc;
	}

	.ptz-up {
		top: 0;
		left: 50px;
	}

	.ptz-left {
		top: 50px;
		left: 0;
	}

	.ptz-right {
		top: 50px;
		left: 100px;
	}

	.ptz-down {
		top: 100px;
		left: 50px;
	}

	.ptz-zoomin {
		top: 150px;
		left: 20px;
	}

	.ptz-talk {
		top: 150px;
		left: 50px;
	}

	.ptz-zoomout {
		top: 150px;
		left: 80px;
	}

	.ptz-up,
	.ptz-left,
	.ptz-right,
	.ptz-down,
	.ptz-center > .fa-microphone,
	.ptz-zoomin,
	.ptz-talk > .fa-microphone,
	.ptz-zoomout {
		cursor: pointer;

		&.readonly {
			cursor: auto;
		}
	}
}

.ptz-block-fs {
	position: fixed;
	margin: 0;
	margin-bottom: 50px;
	top: 0;
	right: 0;
	width: 150px;
	height: 200px;
	text-align: center;
	font-size: 24px;

	.ptz-cell {
		width: 50px;
		height: 50px;
		line-height: 50px;
		position: absolute;
	}

	.ptz-cell.active {
		color: #ccc;
		font-size: 26px;
	}
	.ptz-cell.readonly {
		color: #ddd;
	}
	.fa-microphone, .fa-microphone-slash {
		padding: 10px 13px;
		border-radius: 25px;
	}
	.fa-microphone.active {
		color: #ccc;
	}
	.fa-microphone.readonly {
		color: #ddd;
	}
	.fa-microphone-slash.readonly {
		color: #ddd;
	}
	.mic-level-bar {
		position: absolute;
		transform: rotate(-90deg);
		width: 100%;
		top: 70px;
		left: -80px;
	}

	.ptz-center {
		width: 50px;
		height: 50px;
		top: 50px;
		left: 50px;
		line-height: 50px;
		border-radius: 25px;
		background-color: #ccc;
		cursor: move;
		position: absolute;
	}

	.ptz-up {
		top: 0;
		left: 50px;
	}

	.ptz-left {
		top: 50px;
		left: 0;
	}

	.ptz-right {
		top: 50px;
		left: 100px;
	}

	.ptz-down {
		top: 100px;
		left: 50px;
	}

	.ptz-zoomin {
		top: 150px;
		left: 15px;
	}

	.ptz-zoomout {
		top: 150px;
		left: 85px;
	}

	.ptz-talk {
		top: 150px;
		left: 50px;
	}

	.ptz-up,
	.ptz-left,
	.ptz-right,
	.ptz-down,
	.ptz-center > .fa-microphone,
	.ptz-zoomin,
	.ptz-talk > .fa-microphone,
	.ptz-zoomout {
		cursor: pointer;

		&.readonly {
			cursor: auto;
		}
	}
}

@media screen and(min-width: 768px) {
    .modal-dialog.modal-lgg {
        width: 90%;
    }
    .tab-info {
        display:none;
    }
}

@media screen and(min-width: 1200px) {
    .modal-dialog.modal-lgg {
        width: 1200px;
    }
    .tab-info {
        display:block;
    }
}
</style>
