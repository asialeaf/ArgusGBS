<template>
<FormDlg title="编辑报警预案" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any()">
    <input type="hidden" name="ID" v-model.trim="form.ID">
    <div class="col-md-12">
        <div :class="{'form-group':true, 'has-error': errors.has('Name')}">
            <label class="col-sm-4 control-label">名称
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" name="Name" v-model.trim="form.Name" data-vv-as="名称" v-validate="'required'">
            </div>
        </div>

        <div :class="{'form-group':true, 'has-error': errors.has('Enable')}">
            <label class="col-sm-4 control-label">状态
            </label>
            <div class="col-sm-7 checkbox">
                <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.Enable" name="Enable">
                    启用&nbsp;&nbsp;
                </el-checkbox>
            </div>
        </div>

        <div :class="{'form-group':true, 'has-error': errors.has('SnapInterval')}">
            <label class="col-sm-4 control-label" title="值为0 不截取快照">快照周期(秒)
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" name="SnapInterval" v-model.trim="form.SnapInterval" placeholder="值为0 不截取快照" data-vv-as="快照周期" v-validate="'required|numeric|min_value:0'">
            </div>
        </div>
        <div :class="{'form-group':true, 'has-error': errors.has('RecordDuration')}" v-show="serverInfo.VersionType != '标准版'">
            <label class="col-sm-4 control-label" title="值为0 不录制视频">录像时长(秒)
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" name="RecordDuration" placeholder="值为0 不录制视频" v-model.trim="form.RecordDuration" data-vv-as="录像时长" v-validate="'required|numeric|min_value:0'">
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">报警级别</label>
            <div class="col-sm-7">
                <el-select style="width:100%;" size="medium" v-model.trim="priority" multiple filterable placeholder="全部">
                    <el-option v-for="(item, idx) in priorityList" :key="idx" :label="item.name" :value="item.code">
                    </el-option>
                </el-select>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">报警方式</label>
            <div class="col-sm-7">
                <el-select style="width:100%;" size="medium" v-model.trim="method" multiple filterable placeholder="全部">
                    <el-option v-for="(item, idx) in methodList" :key="idx" :label="item.name" :value="item.code">
                    </el-option>
                </el-select>
            </div>
        </div>
        <div class="form-group" v-if="typeList.length">
            <label class="col-sm-4 control-label">报警类型</label>
            <div class="col-sm-7">
                <el-select style="width:100%;" size="medium" v-model.trim="type" multiple filterable placeholder="全部">
                    <el-option v-for="(item, idx) in typeList" :key="idx" :label="item.name" :value="item.code">
                    </el-option>
                </el-select>
            </div>
        </div>
        <div class="form-group" v-if="eventTypeList.length">
            <label class="col-sm-4 control-label">入侵报警事件类型</label>
            <div class="col-sm-7">
                <el-select style="width:100%;" size="medium" v-model.trim="eventType" multiple filterable placeholder="全部">
                    <el-option v-for="(item, idx) in eventTypeList" :key="idx" :label="item.name" :value="item.code">
                    </el-option>
                </el-select>
            </div>
        </div>
    </div>
    <div class="clearfix"></div>
</FormDlg>
</template>

<script>
import FormDlg from 'components/FormDlg.vue'
import $ from 'jquery'

export default {
    props: {
		serverInfo: {
			type: Object,
			default: () => {}
		},
		userInfo: {
			type: Object,
			default: () => null
		}
    },
    data() {
        return {
            form: this.defForm(),
            priorityList: [{
                "code": "1",
                "name": "一级警情"
            }, {
                "code": "2",
                "name": "二级警情"
            }, {
                "code": "3",
                "name": "三级警情"
            }, {
                "code": "4",
                "name": "四级警情"
            }],
            methodList: [{
                "code": "1",
                "name": "电话报警"
            }, {
                "code": "2",
                "name": "设备报警"
            }, {
                "code": "3",
                "name": "短信报警"
            }, {
                "code": "4",
                "name": "GPS报警"
            }, {
                "code": "5",
                "name": "视频报警"
            }, {
                "code": "6",
                "name": "设备故障报警"
            }, {
                "code": "7",
                "name": "其他报警"
            }],
            priority: [],
            method: [],
            type: [],
            eventType: [],
        }
    },
    computed: {
        eventTypeList() {
            var list = [];
            if(this.method.indexOf("5") >= 0 || !this.method.length) {
                if(this.type.indexOf("5-6") >= 0 || !this.type.length) {
                    list.push({
                        "code": "1",
                        "name": "进入区域"
                    }, {
                        "code": "2",
                        "name": "离开区域"
                    });
                }
            }
            if(!list.length) {
                this.eventType = [];
            }
            return list;
        },
        typeList() {
            var list = [];
            if(this.method.indexOf("2") >= 0 || !this.method.length) {
                list.push({
                    "code": "2-1",
                    "name": "视频丢失报警"
                }, {
                    "code": "2-2",
                    "name": "设备防拆报警"
                }, {
                    "code": "2-3",
                    "name": "存储设备磁盘满报警"
                }, {
                    "code": "2-4",
                    "name": "设备高温报警"
                }, {
                    "code": "2-5",
                    "name": "设备低温报警"
                });
            }
            if(this.method.indexOf("5") >= 0 || !this.method.length) {
                list.push({
                    "code": "5-1",
                    "name": "人工视频报警"
                }, {
                    "code": "5-2",
                    "name": "运动目标检测报警"
                }, {
                    "code": "5-3",
                    "name": "遗留物检测报警"
                }, {
                    "code": "5-4",
                    "name": "物体移除检测报警"
                }, {
                    "code": "5-5",
                    "name": "绊线检测报警"
                }, {
                    "code": "5-6",
                    "name": "入侵检测报警"
                }, {
                    "code": "5-7",
                    "name": "逆行检测报警"
                }, {
                    "code": "5-8",
                    "name": "徘徊检测报警"
                }, {
                    "code": "5-9",
                    "name": "流量统计报警"
                }, {
                    "code": "5-10",
                    "name": "密度检测报警"
                }, {
                    "code": "5-11",
                    "name": "视频异常检测报警"
                }, {
                    "code": "5-12",
                    "name": "快速移动报警"
                }, {
                    "code": "5-13",
                    "name": "图像遮挡报警" // 2022
                });
            }
            if(this.method.indexOf("6") >= 0 || !this.method.length) {
                list.push({
                    "code": "6-1",
                    "name": "存储设备磁盘故障报警"
                }, {
                    "code": "6-2",
                    "name": "存储设备风扇故障报警"
                });
            }
            if(!list.length) {
                this.type = [];
            }
            return list;
        }
    },
    components: {
        FormDlg
    },
    methods: {
        defForm() {
            return {
                ID: "",
                Name: "",
                Enable: false,
                SnapInterval: 30,
                RecordDuration: 15,
                Priority: "",
                Method: "",
                Type: "",
                EventType: "",
            }
        },
        onHide() {
            this.form = this.defForm();
        },
        onShow() {
            this.errors.clear();
        },
        async onSubmit() {
            var ok = await this.$validator.validateAll();
            if (!ok) {
                var e = this.errors.items[0];
                this.$message({
                    type: 'error',
                    message: e.msg
                })
                $(`[name=${e.field}]`).focus();
                return;
            }
            this.form.Priority = this.priority.join(",");
            this.form.Method = this.method.join(",");
            this.form.Type = this.type.join(",");
            this.form.EventType = this.eventType.join(",");
            $.post('/api/v1/alarm/plan/save', {
                ID: this.form.ID,
                Name: this.form.Name,
                SnapInterval: this.form.SnapInterval,
                RecordDuration: this.form.RecordDuration,
                Priority: this.form.Priority,
                Method: this.form.Method,
                Type: this.form.Type,
                EventType: this.form.EventType,
                Enable: this.form.Enable,
            }).then(data => {
                this.$refs['dlg'].hide();
                this.$emit("submit");
            })
        },
        show(data) {
            this.errors.clear();
            if (data) {
                Object.assign(this.form, data);
            }
            if(this.form.Priority) {
                this.priority = this.form.Priority.split(",");
            } else {
                this.priority = [];
            }
            if(this.form.Method) {
                this.method = this.form.Method.split(",");
            } else {
                this.method = [];
            }
            if(this.form.Type) {
                this.type = this.form.Type.split(",");
            } else {
                this.type = [];
            }
            if(this.form.EventType) {
                this.eventType = this.form.EventType.split(",");
            } else {
                this.eventType = [];
            }
            this.$nextTick(() => {
                this.$refs['dlg'].show();
            })
        }
    }
}
</script>

<style lang="less" scoped>
.model-lg {
    width: 80% !important;
}
</style>
