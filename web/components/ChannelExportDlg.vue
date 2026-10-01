<template>
    <FormDlg title="导出通道列表" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any()">
        <div :class="{'form-group':true}" v-if="devid">
            <div class="col-sm-12 checkbox text-center">
                <el-checkbox style="margin-left:-19px;margin-top:-5px;" v-model.trim="downloadThisDev">
                    只导出当前设备({{devid}})的通道
                </el-checkbox>
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('start')}" v-show="showPage">
            <label for="input-start" class="col-sm-3 control-label">开始
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" id="input-start" name="start" v-model.trim="start" data-vv-as="开始" v-validate="'numeric|min_value:0'">
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('limit')}" v-show="showPage">
            <label for="input-limit" class="col-sm-3 control-label">上限
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" id="input-limit" name="limit" v-model.trim="limit" data-vv-as="上限" v-validate="'numeric|min_value:1'">
            </div>
        </div>
    </FormDlg>
</template>

<script>
import FormDlg from 'components/FormDlg.vue';
import $ from 'jquery';

export default {
    props: {
        q: "",
        online: "",
        dir_serial: "",
        channel_type: "",
    },
    data() {
        return {
            debug: false,
            devid: '',
            start: 0,
            limit: 10000,
            downloadThisDev: false,
        }
    },
    components: {
        FormDlg
    },
    beforeDestroy() {
        $(this.$el).off("keydown", this.keyDown);
    },
    computed: {
        showPage() {
            return this.debug || this.start || this.limit != 10000;
        },
        downloadURL() {
            let link = `/api/v1/channel/export?start=${this.start}&limit=${this.limit}`;
            if(this.downloadThisDev && this.devid) {
                link += `&serial=${this.devid}`;
            }
            if(this.q) {
                link += `&q=${this.q}`;
            }
            if(this.online) {
                link += `&online=${this.online}`;
            }
            if(this.dir_serial) {
                link += `&dir_serial=${this.dir_serial}`;
            }
            if(this.channel_type) {
                link += `&channel_type=${this.channel_type}`;
            }
            return link;
        }
    },
    methods: {
        onHide() {
            this.debug = false;
            this.devid = '';
            this.downloadThisDev = false;
            this.start = 0;
            this.limit = 10000;
            this.$emit("hide");
            $(this.$el).off("keydown", this.keyDown);
        },
        onShow() {
            this.errors.clear();
            this.$emit("show");
            $(this.$el).on("keydown", this.keyDown);
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
        async onSubmit() {
            var ok = await this.$validator.validateAll();
            if(!ok) {
                var e = this.errors.items[0];
                this.$message({
                    type: 'error',
                    message: e.msg
                })
                $(`[name=${e.field}]`).focus();
                return;
            }
            window.open(this.downloadURL, "_blank");
            this.$refs['dlg'].hide();
            this.$emit("submit");
        },
        show(devid) {
            this.errors.clear();
            this.devid = devid;
            this.$nextTick(() => {
                this.$refs['dlg'].show();
            })
        }
    }
}
</script>
