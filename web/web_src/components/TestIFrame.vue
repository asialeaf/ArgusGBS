<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-primary" @click.prevent="send" :disabled="sending">发送<span v-show="sending">...</span></button>
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

            $.get("/api/v1/control/iframe", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "强制关键帧发送成功"
                })
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).always(() => {
                this.sending = false;
            })
        },
    }
}
</script>

<style lang="less" scoped>
.play-area {
    margin: 20px auto;
}
</style>
