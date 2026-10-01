<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <label class="col-sm-4 control-label">报警方式
            </label>
            <div class="col-sm-7">
                <select class="form-control" v-model.trim="alarm_method">
                    <option value="">不选</option>
                    <option value="1">1-电话报警</option>
                    <option value="2">2-设备报警</option>
                    <option value="3">3-短信报警</option>
                    <option value="4">4-GPS报警</option>
                    <option value="5">5-视频报警</option>
                    <option value="6">6-设备故障报警</option>
                    <option value="7">7-其他报警</option>
                </select>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">报警类型
            </label>
            <div class="col-sm-7">
                <select class="form-control" v-model.trim="alarm_type">
                    <option v-for="type in alarmTypes" :key="type.value" :value="type.value">{{ type.label }}</option>
                </select>
            </div>
        </div>
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-primary" @click.prevent="send" :disabled="sending">报警复位<span v-show="sending">...</span></button>
            </div>
        </div>
    </div>
</div>
</template>

<script>

export default {
    data() {
        return {
            sending: false,
            alarm_method: "",
            alarm_type: "",
        }
    },
    computed: {
        alarmTypes() {
            var list = [{ value: "", label: "不选" }];
            switch(this.alarm_method) {
                case "2": // 设备报警
                    list.push({ value: "1", label: "1-视频丢失报警" },{ value: "2", label: "2-设备防拆报警" },
                        { value: "3", label: "3-存储设备磁盘满报警" }, { value: "4", label: "4-设备高温报警" },
                        { value: "5", label: "5-设备低温报警" });
                    break;
                case "5": // 视频报警
                    list.push({ value: "1", label: "1-人工视频报警" },{ value: "2", label: "2-运动目标检测报警" },
                        { value: "3", label: "3-遗留物检测报警" }, { value: "4", label: "4-物体移除检测报警" },
                        { value: "5", label: "5-绊线检测报警" }, { value: "6", label: "6-入侵检测报警" },
                        { value: "7", label: "7-逆行检测报警" }, { value: "8", label: "8-徘徊检测报警" },
                        { value: "9", label: "9-流量统计报警" }, { value: "10", label: "10-密度检测报警" },
                        { value: "11", label: "11-视频异常检测报警" }, { value: "12", label: "12-快速移动报警" },
                        { value: "13", label: "13-图像遮挡报警" });
                    break;
                case "6": // 设备故障报警
                    list.push({ value: "1", label: "1-存储设备磁盘故障报警" },{ value: "2", label: "2-存储设备风扇故障报警" });
                    break;
            }
            return list;
        },
    },
    watch: {
        alarm_method: function(newVal, oldVal) {
            this.alarm_type = "";
        },
    },
    methods: {
        async send() {
            this.sending = true;
            await this.$store.dispatch("connect");

            $.post("/api/v1/control/alarmreset", {
                serial: this.$store.state.serial,
                code: this.$store.state.code||this.$store.state.serial,
                alarm_method: this.alarm_method,
                alarm_type: this.alarm_type,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "报警复位操作成功"
                })
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).fail(ret => {
                this.$store.commit("updateResult", null);
            }).always(() => {
                this.sending = false;
            })
        },
    },
}
</script>
