<template>
    <div :class="['modal', { fade: fade }]" data-keyboard="true" data-backdrop="static" tabindex="-1">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <button type="button" class="close" data-dismiss="modal" aria-label="Close">
					    <span aria-hidden="true">&times;</span>
					</button>
                    <h4 class="modal-title text-primary text-center"><span>{{videoTitle}}</span></h4>
                </div>
                <div class="modal-body">
                    <LivePlayer v-if="bShow" :videoUrl="videoUrl" :waterMark="osd" :snapUrl="snapUrl"
                        @media_info="onMediaInfo" @error="onError" @message="$message"
                        :digitalZoom="digitalZoom" :hideBigPlayButton="!!serverInfo.HideBigPlayButton"
                        v-loading="bLoading" :loading.sync="bLoading" element-loading-text="加载中..." element-loading-background="#000">
                    </LivePlayer>
                    <div class="text-center" v-if="isDemoUser(serverInfo, userInfo) && !bOutHevcTip">
                        <br>
                        提示: 演示系统限制匿名登录播放时间, 若需测试长时间播放, 请<a target="_blank" href="//www.liveqing.com/docs/download/LiveGBS.html">下载使用</a>
                    </div>
                    <div class="text-center text-red" v-if="bOutHevcTip">
                        <br>
                        提示: 正在播放 H265 直出流, 确保浏览器版本较新, 并且开启硬件加速
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-default" data-dismiss="modal">关闭</button>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import 'jquery-ui/ui/widgets/draggable'
import LivePlayer from '@liveqing/liveplayer'

export default {
    data() {
        return {
            videoUrl: "",
            videoTitle: "",
            osd: "",
            snapUrl: "",
            mediaInfo: null,
            bShow: false,
            bLoading: false,
            bOutHevcTip: false,
        }
    },
    props: {
        fade: {
            type: Boolean,
            default: false
        },
        digitalZoom: {
            type: Boolean,
            default: true
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
    mounted() {
        $(this.$el).find('.modal-content').draggable({
            handle: '.modal-header',
            cancel: ".modal-title span",
            addClasses: false,
            containment: 'document',
            delay: 100,
            opacity: 0.5
        });
        $(this.$el).on("hide.bs.modal", () => {
            this.reset();
            this.$nextTick(() => {
                this.bShow = false;
            })
        }).on("show.bs.modal", () => {
            this.bShow = true;
        })
    },
    components: { LivePlayer },
    methods: {
        reset() {
            this.videoUrl = "";
            this.videoTitle = "";
            this.osd = "";
            this.snapUrl = "";
            this.mediaInfo = null;
            this.bLoading = false;
            this.bOutHevcTip = false;
        },
        play(src, title, snap, osd) {
            this.videoUrl = src || "";
            this.videoTitle = title || "";
            this.snapUrl = snap || "";
            this.osd = osd || "";

            $(this.$el).modal("show");
        },
        onMediaInfo(mi) {
            this.mediaInfo = mi;
        },
        onError(e) {
            if(e == 'MediaError' && this.mediaInfo && String(this.mediaInfo.videoCodec).startsWith("hvc")) {
                if(flvjs.getFeatureList() && !flvjs.getFeatureList().nativeMP4H265Playback) {
                    this.bOutHevcTip = true;
                    console.log("提示: 正在播放 H265 直出流, 确保浏览器版本较新, 并且开启硬件加速");
                }
            }
        },
    }
}
</script>

<style lang="less" scoped>
.modal-title {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
}
</style>
