<template>
    <div :class="['modal', { fade: fade }]" data-keyboard="true" data-backdrop="static" tabindex="-1">
        <div class="modal-dialog modal-md">
            <div class="modal-content">
                <div class="modal-header">
                    <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                        <span aria-hidden="true">&times;</span>
                    </button>
                    <h4 class="modal-title text-primary text-center"><span>{{title}}</span></h4>
                </div>
                <div class="modal-body">
                    <el-progress :text-inside="true" :stroke-width="18" :percentage="progress"></el-progress>
                    <p v-if="fileError" class="text-red text-center">
                        <br>
                        <span class="text-bold">异常终止提示：</span>
                        {{fileError}}
                    </p>
                </div>
                <div class="modal-footer">
					<span v-show="inBitRate">{{inBitRate}}Kbps&nbsp;&nbsp;</span>
					<el-radio-group id="speed-switcher-download" size="small" v-model.number="speed" @change="scale" v-show="fileUrl">
						<el-radio-button label="0.5">0.5x</el-radio-button>
						<el-radio-button label="1">1x</el-radio-button>
						<el-radio-button label="2">2x</el-radio-button>
						<el-radio-button label="4">4x</el-radio-button>
					</el-radio-group>
                    <button type="button" class="btn btn-primary" @click.prevent="hide" v-show="fileUrl">下载</button>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import "jquery-ui/ui/widgets/draggable";

export default {
	data() {
		return {
			bShow: false,
			timer: 0,
			title: "",
			streamID: "",
			progress: 0,
			speed: 1,
			inBitRate: 0,
			fileUrl: "",
			fileError: "",
		};
	},
	props: {
		fade: {
			type: Boolean,
			default: false
		},
	},
	mounted() {
		$(this.$el).find(".modal-content").draggable({
			handle: ".modal-header",
			cancel: ".modal-title span",
			addClasses: false,
			containment: "document",
			delay: 100,
			opacity: 0.5
		});
		$(this.$el).on("shown.bs.modal", () => {
			this.bShow = true;
			this.$emit("show");
			if(this.streamID) {
				this.timer = setInterval(() => {
					$.ajax({
						type: "GET",
						url: "/api/v1/playback/streaminfo",
						data: {
                            streamid: this.streamID
						},
						global: false
					}).then(ret => {
						this.progress = Math.ceil(ret.Progress * 100);
						this.fileUrl = ret.PlaybackFileURL || "";
						this.fileError = ret.PlaybackFileError || "";
						this.speed = ret.PlaybackSpeed || 1;
						this.inBitRate = ret.InBitRate || 0;
					}).fail(() => {
						this.progress = 100;
						this.inBitRate = 0;
					})
				}, 3000);
			}
		}).on("hidden.bs.modal", () => {
			this.bShow = false;
			this.$emit("hide");
			if(this.timer) {
				clearInterval(this.timer);
				this.timer = 0;
			}
			this.stop();
		})
	},
	beforeDestroy() {
		if (this.timer) {
			clearInterval(this.timer);
			this.timer = 0;
		}
	},
	methods: {
		show() {
			$(this.$el).modal("show");
		},
		hide() {
			this.stop(true);
			$(this.$el).modal("hide");
		},
		doSubmit() {
			this.$emit("submit");
		},
		download(title, streamID) {
			this.title = title;
			this.streamID = streamID;
			this.fileUrl = "";
			this.fileError = "";
			this.progress = 0;
			this.speed = 1;
			this.inBitRate = 0;
			$(this.$el).modal("show");
		},
		scale(speed) {
			if(!this.streamID) return;
			$.post("/api/v1/playback/control", {
				streamid: this.streamID,
				command: "scale",
				scale: speed,
			}).then(data => {
				this.$message({
					type: "success",
					message: `${speed} 倍速设置成功`,
				});
			})
		},
		stop(download = false) {
			if(this.streamID) {
				$.ajax({
					type: "POST",
					url: "/api/v1/playback/stop",
					data: {
						streamid: this.streamID
					},
					global: false
				}).always(() => {
					if(this.fileUrl && download) {
						window.open(this.fileUrl, "_blank");
					}
					this.fileUrl = "";
					this.fileError = "";
					this.progress = 0;
					this.inBitRate = 0;
					this.$emit("download");
				})
				this.streamID = "";
			}
		},
	}
};
</script>

<style lang="less" scoped>
.modal-title {
	overflow: hidden;
	white-space: nowrap;
	text-overflow: ellipsis;
}

#speed-switcher-download {
	margin-right: 10px;
	label {
		margin-bottom: 0;
	}
}
</style>