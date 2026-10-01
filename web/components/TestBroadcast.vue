<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group" v-show="debug">
            <div class="col-sm-offset-4 col-sm-7 checkbox">
                <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model="byeOnClose">
                    断开时立即结束会话&nbsp;&nbsp;
                </el-checkbox>
            </div>
        </div>
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-primary" @click.prevent="start" :disabled="starting||!!ws">开始广播<span v-show="starting">...</span></button>
                <button role="button" class="btn btn-danger" @click.prevent="stop" :disabled="stopping" v-if="!!ws">停止广播<span v-show="stopping">...</span></button>
                <span v-show="ws" style="vertical-align: middle; color: #666;">
                    &nbsp;&nbsp;
                    <i :class="['fa fa-2x', {
                        'fa-volume-off': level <= 0,
                        'fa-volume-down': level > 0 && level < 30,
                        'fa-volume-up': level >= 30,
                    }]"></i>
                </span>
            </div>
        </div>
    </div>
    <div class="play-area" v-if="videoUrl">
        <div class="col-sm-10 col-sm-offset-1">
            <LivePlayer ref="player" :videoUrl="videoUrl" live stretch :muted="!audio" :controls="false">
            </LivePlayer>
        </div>
    </div>
</div>
</template>

<script>
import LivePlayer from "@liveqing/liveplayer";

export default {
    data() {
        return {
            starting: false,
            stopping: false,
            recorder: null,
            ws: null,
            audio: false,
            videoUrl: "",
            level: 0,
            debug: false,
            byeOnClose: true
        }
    },
    components: { LivePlayer },
    methods: {
        async start() {
            if(!this.$store.state.serial || !this.$store.state.code) {
                this.$message({
                    type: "error",
                    message: "请先选择设备通道"
                })
                return;
            }
            var streamInfo = await this.$store.dispatch("play");
            if(streamInfo) {
                var videoUrl = streamInfo.HLS || "";
                if(this.flvSupported()) {
                    videoUrl = this.isIE() ? streamInfo.WS_FLV || "" : streamInfo.FLV || "";
                }
                this.audio = !!streamInfo.AudioEnable;
                this.videoUrl = videoUrl;
            }

            this.starting = true;
            await this.$store.dispatch("connect");

			if(!this.ws) {
                try {
                    this.ws = await this.doConnect();
                } catch (error) {
                    this.$message({
                        type: "error",
                        message: `开始广播失败`
                    })
                    console.log(error)
                    this.starting = false;
                    return
                }
			}
            this.starting = false;
			if(this.recorder) {
				this.recorder.start();
				return;
			}
			LiveRecorder.get((rec, err) => {
				if(err) {
					alert(err);
					this.stop();
					return
				}
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
        stop() {
			if(this.recorder) {
				// this.recorder.stop();
				this.recorder.destroy();
				this.recorder = null;
			}
			if(this.ws) {
				this.ws.close();
				this.ws = null;
			}
			this.level = 0;
            setTimeout(() => {
                this.$store.dispatch("disconnect");
            }, this.byeOnClose ? 1000 : 11000);
        },
        doConnect() {
            return new Promise((resolve, reject) => {
                const ws = new WebSocket(this.wsTalkURL());
                ws.onopen = evt => {
					resolve(ws);
				}
				ws.onerror = evt => {
					reject(evt);
				}
            })
        },
		wsTalkURL() {
			var protocol = "ws:";
			if(location.protocol.startsWith("https")) {
				protocol = "wss:";
			}
			//url query param "format=pcm|g711a", "bye_on_close=0|1"
			return `${protocol}//${location.host}/api/v1/control/ws-talk/${this.$store.state.serial}/${this.$store.state.code}?format=pcm&bye_on_close=${this.byeOnClose}`;
		},
        keyDown(e) {
            if(e.altKey && e.shiftKey) {
                switch(e.key) {
                    case 'D':
                        e.preventDefault();
                        this.toggleDebug();
                        break;
                }
            }
        },
        toggleDebug() {
            this.debug = !this.debug;
        },
    },
    mounted() {
        $(document).on("keydown", this.keyDown);
    },
    beforeDestroy() {
        $(document).off("keydown", this.keyDown);
    },
}
</script>

<style lang="less" scoped>
.play-area {
    margin: 50px 0 10px 0;
}
</style>
