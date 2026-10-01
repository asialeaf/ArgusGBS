<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <label class="col-sm-4 control-label">Pan(0~360)
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-input-number size="small" v-model="pan" :min="0" :max="360" :precision="1"></el-input-number>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">Tilt(-30~90)
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-input-number size="small" v-model="tilt" :min="-30" :max="90" :precision="1"></el-input-number>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">Zoom
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-input-number size="small" v-model="zoom" :min="1" :precision="1"></el-input-number>
            </div>
        </div>
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-primary" @click.prevent="send" :disabled="sending">精准PTZ控制<span v-show="sending">...</span></button>
                <button role="button" class="btn btn-info" @click.prevent="query" :disabled="loading">查询<span v-show="loading">...</span></button>
            </div>
        </div>
    </div>

    <div class="play-area" v-if="videoUrl">
        <div class="row">
            <div class="col-sm-10 col-sm-offset-1 col-lg-8 col-lg-offset-2">
                <LivePlayer ref="player" :videoUrl="videoUrl" live stretch :muted="!audio" :controls="false">
                </LivePlayer>
            </div>
        </div>
    </div>
</div>
</template>

<script>
import LivePlayer from "@liveqing/liveplayer";

export default {
    data() {
        return {
            sending: false,
            loading: false,
            pan: 264.3,
            tilt: 0.0,
            zoom: 10.3,
            videoUrl: "",
            audio: false,
        }
    },
    components: { LivePlayer },
    methods: {
        async send() {
            var streamInfo = await this.$store.dispatch("play");
            if(streamInfo) {
                var videoUrl = streamInfo.HLS || "";
                if(this.flvSupported()) {
                    videoUrl = this.isIE() ? streamInfo.WS_FLV || "" : streamInfo.FLV || "";
                }
                this.audio = !!streamInfo.AudioEnable;
                this.videoUrl = videoUrl;
            }

            this.sending = true;
            await this.$store.dispatch("connect");

            $.post("/api/v1/control/ptzprecise", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
                pan: this.pan,
                tilt: this.tilt,
                zoom: this.zoom,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "精准PTZ控制成功"
                })
            }).always(() => {
                this.$store.commit("updateResult", null);
                this.sending = false;
            })
        },
        async query() {
            this.loading = true;
            await this.$store.dispatch("connect", {
                topic: "精准PTZ查询",
            });

            $.get("/api/v1/device/fetchptzposition", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "精准PTZ查询成功"
                })
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).always(() => {
                this.loading = false;
            })
        },
    },
}
</script>

<style lang="less" scoped>
.el-radio-group {
    margin-top: 9px;

    .el-radio + .el-radio {
        margin-left: 15px;
    }
}

.play-area {
    margin: 20px auto;
}
</style>
