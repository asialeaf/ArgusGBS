<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <label class="col-sm-4 control-label">水平速度</label>
            <div class="col-sm-7">
                <el-slider v-model.number="speedx" :min="0" :max="255"></el-slider>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">垂直速度</label>
            <div class="col-sm-7">
                <el-slider v-model.number="speedy" :min="0" :max="255"></el-slider>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">变倍速度</label>
            <div class="col-sm-7">
                <el-slider v-model.number="speedz" :min="0" :max="15"></el-slider>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">光圈速度</label>
            <div class="col-sm-7">
                <el-slider v-model.number="speedi" :min="0" :max="255"></el-slider>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">聚焦速度</label>
            <div class="col-sm-7">
                <el-slider v-model.number="speedf" :min="0" :max="255"></el-slider>
            </div>
        </div>
    </div>
    <div class="control-area">
        <div class="row">
            <div class="col-sm-4">
                <div class="control-row">
                    <el-button size="mini" :disabled="ptzing" @click="ptz('leftup')"><i class="fa fa-lg fa-caret-up" style="transform: rotate(-45deg);"></i></el-button>
                    <el-button size="mini" :disabled="ptzing" @click="ptz('up')"><i class="fa fa-lg fa-caret-up"></i></el-button>
                    <el-button size="mini" :disabled="ptzing" @click="ptz('rightup')"><i class="fa fa-lg fa-caret-up" style="transform: rotate(45deg);"></i></el-button>
                </div>
                <div class="control-row">
                    <el-button size="mini" :disabled="ptzing" @click="ptz('left')"><i class="fa fa-lg fa-caret-left"></i></el-button>
                    <el-button size="mini" :disabled="stopping" @click="stop"><i class="fa fa-stop text-danger"></i></el-button>
                    <el-button size="mini" :disabled="ptzing" @click="ptz('right')"><i class="fa fa-lg fa-caret-right"></i></el-button>
                </div>
                <div class="control-row">
                    <el-button size="mini" :disabled="ptzing" @click="ptz('leftdown')"><i class="fa fa-lg fa-caret-left" style="transform: rotate(-45deg);"></i></el-button>
                    <el-button size="mini" :disabled="ptzing" @click="ptz('down')"><i class="fa fa-lg fa-caret-down"></i></el-button>
                    <el-button size="mini" :disabled="ptzing" @click="ptz('rightdown')"><i class="fa fa-lg fa-caret-right" style="transform: rotate(45deg);"></i></el-button>
                </div>
            </div>
            <div class="col-sm-4">
                <div class="control-row">
                    <span>变倍:</span>
                    <el-button size="mini" circle icon="el-icon-minus" :disabled="zooming" @click="zoom('zoomout')"></el-button>
                    <el-button size="mini" circle icon="el-icon-plus" :disabled="zooming" @click="zoom('zoomin')"></el-button>
                    <el-button size="mini" circle :disabled="stopping" @click="stop"><i class="fa fa-stop text-danger"></i></el-button>
                </div>
                <div class="control-row">
                    <span>变焦:</span>
                    <el-button size="mini" circle icon="el-icon-minus" :disabled="focusing" @click="focus('focusfar')"></el-button>
                    <el-button size="mini" circle icon="el-icon-plus" :disabled="focusing" @click="focus('focusnear')"></el-button>
                    <el-button size="mini" circle :disabled="stopping" @click="stop"><i class="fa fa-stop text-danger"></i></el-button>
                </div>
                <div class="control-row">
                    <span>光圈:</span>
                    <el-button size="mini" circle icon="el-icon-minus" :disabled="irising" @click="iris('irisout')"></el-button>
                    <el-button size="mini" circle icon="el-icon-plus" :disabled="irising" @click="iris('irisin')"></el-button>
                    <el-button size="mini" circle :disabled="stopping" @click="stop"><i class="fa fa-stop text-danger"></i></el-button>
                </div>
            </div>
            <div class="col-sm-4">
                <el-radio-group class="control-drag-zoom" v-model.trim="dragCmd">
                    <el-radio label="zoomin">拉框放大</el-radio>
                    <el-radio label="zoomout">拉框缩小</el-radio>
                    <el-radio label="">关闭拉框</el-radio>
                </el-radio-group>
            </div>
        </div>
    </div>
    <div class="play-area" v-if="videoUrl">
        <div class="row">
            <div class="col-sm-10 col-sm-offset-1 col-lg-8 col-lg-offset-2"
                @mousedown="startDrawing" @touchstart="startDrawing"
                @mousemove="draw" @touchmove="draw"
                @mouseup="stopDrawing($event, true)" @touchend="stopDrawing($event, true)"
                @mouseleave="stopDrawing($event, false)">
                <LivePlayer ref="player" :videoUrl="videoUrl" live stretch :muted="!audio" :controls="false">
                    <div v-if="isDrawing" class="rectangle" :style="{
                        left: rect.x + 'px',
                        top: rect.y + 'px',
                        width: rect.width + 'px',
                        height: rect.height + 'px'
                    }"></div>
                </LivePlayer>
            </div>
        </div>
    </div>
</div>
</template>

<script>
import { mapState } from "vuex";
import LivePlayer from "@liveqing/liveplayer";

export default {
    data() {
        return {
            speedx: 155,
            speedy: 155,
            speedz: 10,
            speedi: 155,
            speedf: 155,
            dragCmd: "",

            ptzing: false,
            zooming: false,
            focusing: false,
            irising: false,
            draging: false,
            stopping: false,
            audio: false,
            videoUrl: "",

            isDrawing: false,
            startPos: { x: 0, y: 0 },
            rect: { x: 0, y: 0, width: 0, height: 0 },
        }
    },
    computed: {
        ...mapState( ["serial", "code"]),
    },
    watch: {
        serial: function(newVal, oldVal) {
            if(this.videoUrl) this.videoUrl = "";
        },
        code: function(newVal, oldVal) {
            if(this.videoUrl) this.videoUrl = "";
        },
    },
    components: { LivePlayer },
    methods: {
        async ptz(command) {
            if(!command) return;
            this.ptzing = true;
            if(!this.videoUrl) {
                var streamInfo = await this.$store.dispatch("play");
                if(streamInfo) {
                    var videoUrl = streamInfo.HLS || "";
                    if(this.flvSupported()) {
                        videoUrl = this.isIE() ? streamInfo.WS_FLV || "" : streamInfo.FLV || "";
                    }
                    this.audio = !!streamInfo.AudioEnable;
                    this.videoUrl = videoUrl;
                }
            }
            await this.$store.dispatch("connect");

            $.post("/api/v1/control/ptz", {
				serial: this.$store.state.serial,
				code: this.$store.state.code,
                command: command,
                speedx: this.speedx,
                speedy: this.speedy,
            }).always(() => {
                this.ptzing = false;
            })
        },
        async zoom(command) {
            if(!command) return;
            this.zooming = true;
            if(!this.videoUrl) {
                var streamInfo = await this.$store.dispatch("play");
                if(streamInfo) {
                    var videoUrl = streamInfo.HLS || "";
                    if(this.flvSupported()) {
                        videoUrl = this.isIE() ? streamInfo.WS_FLV || "" : streamInfo.FLV || "";
                    }
                    this.audio = !!streamInfo.AudioEnable;
                    this.videoUrl = videoUrl;
                }
            }
            await this.$store.dispatch("connect");

            $.post("/api/v1/control/ptz", {
				serial: this.$store.state.serial,
				code: this.$store.state.code,
                command: command,
                speed: this.speedz,
            }).always(() => {
                this.zooming = false;
            })
        },
        async focus(command) {
            if(!command) return;
            this.focusing = true;
            if(!this.videoUrl) {
                var streamInfo = await this.$store.dispatch("play");
                if(streamInfo) {
                    var videoUrl = streamInfo.HLS || "";
                    if(this.flvSupported()) {
                        videoUrl = this.isIE() ? streamInfo.WS_FLV || "" : streamInfo.FLV || "";
                    }
                    this.audio = !!streamInfo.AudioEnable;
                    this.videoUrl = videoUrl;
                }
            }
            await this.$store.dispatch("connect");

            $.post("/api/v1/control/fi", {
				serial: this.$store.state.serial,
				code: this.$store.state.code,
                command: command,
                speed: this.speedf,
            }).always(() => {
                this.focusing = false;
            })
        },
        async iris(command) {
            if(!command) return;
            this.irising = true;
            if(!this.videoUrl) {
                var streamInfo = await this.$store.dispatch("play");
                if(streamInfo) {
                    var videoUrl = streamInfo.HLS || "";
                    if(this.flvSupported()) {
                        videoUrl = this.isIE() ? streamInfo.WS_FLV || "" : streamInfo.FLV || "";
                    }
                    this.audio = !!streamInfo.AudioEnable;
                    this.videoUrl = videoUrl;
                }
            }
            await this.$store.dispatch("connect");

            $.post("/api/v1/control/fi", {
				serial: this.$store.state.serial,
				code: this.$store.state.code,
                command: command,
                speed: this.speedi,
            }).always(() => {
                this.irising = false;
            })
        },
        async drag(command) {
            if(!command) return;
            const bg = this.$el.querySelector("video").getBoundingClientRect();
            const bgw = Math.ceil(bg.width);
            const bgh = Math.ceil(bg.height);
            const midx = Math.ceil(this.rect.x+this.rect.width/2);
            const midy = Math.ceil(this.rect.y+this.rect.height/2);
            const w = Math.ceil(this.rect.width);
            const h = Math.ceil(this.rect.height);
            if(!bgw || !bgh || !midx || !midy || !w || !h) return;

            this.draging = true;
            if(!this.videoUrl) {
                var streamInfo = await this.$store.dispatch("play");
                if(streamInfo) {
                    var videoUrl = streamInfo.HLS || "";
                    if(this.flvSupported()) {
                        videoUrl = this.isIE() ? streamInfo.WS_FLV || "" : streamInfo.FLV || "";
                    }
                    this.audio = !!streamInfo.AudioEnable;
                    this.videoUrl = videoUrl;
                }
            }
            await this.$store.dispatch("connect");

            $.post(`/api/v1/control/drag${command}`, {
				serial: this.$store.state.serial,
				code: this.$store.state.code,
                length: bgw,
                width: bgh,
                midpointx: midx,
                midpointy: midy,
                lengthx: w,
                lengthy: h,
            }).always(() => {
                this.draging = false;
            })
        },
        stop() {
            this.stopping = true;
            $.post("/api/v1/control/ptz", {
				serial: this.$store.state.serial,
				code: this.$store.state.code,
                command: "stop",
            }).then(() => {
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).always(() => {
                this.stopping = false;
            })
        },
        startDrawing(event) {
            if(!this.dragCmd) return;
            event.preventDefault();
            this.isDrawing = true;
            this.startPos = this.getPosition(event);
            this.rect = { ...this.startPos, width: 0, height: 0 };
        },
        draw(event) {
            if (!this.isDrawing) return
            event.preventDefault()
            const currentPos = this.getPosition(event)

            this.rect = {
                x: Math.min(this.startPos.x, currentPos.x),
                y: Math.min(this.startPos.y, currentPos.y),
                width: Math.abs(currentPos.x - this.startPos.x),
                height: Math.abs(currentPos.y - this.startPos.y)
            }
        },
        stopDrawing(event, drag) {
            this.isDrawing = false;
            if(drag && this.dragCmd) {
                this.drag(this.dragCmd);
            }
        },
        getPosition(event) {
            const rect = this.$el.querySelector("video").getBoundingClientRect();
            return {
                x: (event.clientX || event.touches[0].clientX) - rect.left,
                y: (event.clientY || event.touches[0].clientY) - rect.top
            }
        },
    },
}
</script>

<style lang="less" scoped>
@import url(~assets/styles/variables.less);

.control-area {
    > .row {
        > .col-sm-4 {
            margin: 10px 0;
        }
    }
}
.control-row {
    text-align: center;
    white-space: pre;

    .el-button--mini {
        display: inline-block;
        width: 35px;
        height: 35px;
        padding: 10px;
    }
    .el-button + .el-button {
        margin-left: 5px;
    }

    + .control-row {
        margin-top: 5px;
    }
}

.control-drag-zoom {
    display: block;

    .el-radio {
        display: block;
        margin: 5px auto;
        max-width: 100px;
    }
}

.play-area {
    margin: 20px auto;

    .rectangle {
        position: absolute;
        pointer-events: none;
        background-color: rgba(0, 255, 0, 0.3);
    }
}
</style>
