import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import * as _ from 'lodash';
import EditorConfig from 'src/app/@models/editorconfig.model';
import { ApiService } from '../../@services/api.service';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ModalYamlPreviewComponent } from 'src/app/@modals/modal-yaml-preview/modal-yaml-preview.component';
import Template from 'src/app/@models/template.model';
import { ResolutionService } from 'src/app/@services/resolution.service';
import { RequestService } from 'src/app/@services/request.service';
import MetaUtask from 'src/app/@models/meta-utask.model';
import Task from '../../@models/task.model';
import { TaskService } from 'src/app/@services/task.service';
import { ActiveIntervalService } from 'src/app/@services/active-interval.service';

@Component({
  templateUrl: './task.html'
})
export class TaskComponent implements OnInit, OnDestroy {
  objectKeys = Object.keys;
  loaders: { [key: string]: boolean } = {};
  errors: { [key: string]: any } = {};
  display: { [key: string]: boolean } = {};
  confirm: { [key: string]: boolean } = {};
  refreshes: { [key: string]: any } = {};
  item: any = {
    resolver_inputs: {},
    task_id: null
  };
  task: Task = null;
  taskIsResolvable = false;
  autorefresh = false;
  taskId = '';
  resolution: any = null;
  // steps: any[] = [];
  editorConfigResult: EditorConfig = {
    readonly: true,
    mode: 'ace/mode/json',
    theme: 'ace/theme/monokai',
    maxLines: 25,
  };
  selectedStep = '';
  meta: MetaUtask = null;

  JSON = JSON;
  template: Template;
  comment: any = {
    content: ''
  };

  constructor(private modalService: NgbModal, private api: ApiService, private route: ActivatedRoute, private resolutionService: ResolutionService, private requestService: RequestService, private taskService: TaskService, private router: Router) {
  }

  ngOnDestroy() {
    this.refreshes.tasks.stopInterval();
  }

  ngOnInit() {
    this.meta = this.route.parent.snapshot.data.meta;
    this.route.params.subscribe(params => {
      this.taskId = params.id;
      this.loadTask().then(() => {
        this.display.request = (!this.task.result && !this.resolution) || (!this.resolution && this.taskIsResolvable);
        this.display.result = !!this.task.result;
        this.display.execution = !!this.resolution;
        this.display.reject = !this.resolution && this.taskIsResolvable;
        this.display.resolution = !this.resolution && this.taskIsResolvable;
        this.display.comments = this.task.comments && this.task.comments.length > 0;
        this.autorefresh = ['TODO', 'RUNNING'].indexOf(this.task.state) > -1;
      });
    });

    this.refreshes.tasks = new ActiveIntervalService();
    this.refreshes.tasks.setInterval(() => {
      if (!this.loaders.tasks && this.autorefresh) {
        this.loadTask().then(() => {
          console.log('refresh');
        });
      }
    }, 15000, false);
  }

  addComment() {
    this.loaders.addComment = true;
    this.api.addComment(this.task.id, this.comment.content).toPromise().then((comment) => {
      this.task.comments = this.task.comments ? this.task.comments : [];
      this.task.comments.push(comment);
      this.errors.addComment = null;
      this.comment.content = '';
    }).catch((err) => {
      this.errors.addComment = err;
    }).finally(() => {
      this.loaders.addComment = false;
    });
  }

  previewDetails(obj: any, title: string) {
    const previewModal = this.modalService.open(ModalYamlPreviewComponent, {
      size: 'xl'
    });
    previewModal.componentInstance.value = obj;
    previewModal.componentInstance.title = title;
  }

  editRequest(task: any) {
    this.requestService.edit(task).then((data: any) => {
      this.loadTask();
    }).catch((err) => {
      if (err !== 0) {
        console.log(err);
      }
    });
  }

  editResolution(resolution: any) {
    this.resolutionService.edit(resolution).then((data: any) => {
      console.log(data);
      this.loadTask();
    }).catch((err) => {
      if (err !== 0) {
        console.log(err);
      }
    });
  }

  runResolution(resolution: any) {
    this.resolutionService.run(resolution.id).then((data: any) => {
      console.log(data);
      this.loadTask();
    }).catch((err) => {
      if (err !== 0) {
        console.log(err);
      }
    });
  }

  pauseResolution(resolution: any) {
    this.resolutionService.pause(resolution.id).then((data: any) => {
      console.log(data);
      this.loadTask();
    }).catch((err) => {
      if (err !== 0) {
        console.log(err);
      }
    });
  }

  cancelResolution(resolution: any) {
    this.resolutionService.cancel(resolution.id).then((data: any) => {
      console.log(data);
      this.loadTask();
    }).catch((err) => {
      if (err !== 0) {
        console.log(err);
      }
    });
  }

  extendResolution(resolution: any) {
    this.resolutionService.extend(resolution.id).then((data: any) => {
      console.log(data);
      this.loadTask();
    }).catch((err) => {
      if (err !== 0) {
        console.log(err);
      }
    });
  }

  deleteTask(taskId: string) {
    this.taskService.delete(taskId).then((data: any) => {
      this.router.navigate([`/home`]);
    }).catch((err) => {
      if (err !== 0) {
        console.log(err);
      }
    });
  }

  rejectTask() {
    this.loaders.rejectTask = true;
    this.api.rejectTask(this.task.id).toPromise().then((res: any) => {
      console.log('OK', res);
      this.loadTask();
    }).catch((err) => {
      console.log('ERROR', err);
    }).finally(() => {
      this.loaders.rejectTask = false;
    });
  }

  resolveTask() {
    this.loaders.resolveTask = true;
    this.api.postResolution(this.item).toPromise().then((res: any) => {
      console.log('OK', res);
      this.loadTask();
    }).catch((err) => {
      console.log('ERROR', err);
    }).finally(() => {
      this.loaders.resolveTask = false;
    });
  }

  selectStepFromViewer(step) {
    this.selectedStep = step || '';
  }

  // generateSteps() {
  //   const steps = [];
  //   if (
  //     _.get(this, 'resolution.steps', null) &&
  //     _.isObjectLike(this.resolution.steps)
  //   ) {
  //     _.each(this.resolution.steps, (data: any, key: string) => {
  //       steps.push({ key, data });
  //     });
  //     this.steps = steps;
  //   } else {
  //     this.steps = [];
  //   }
  // }

  loadTask() {
    return new Promise((resolve, reject) => {
      this.loaders.task = true;
      this.api.task(this.taskId).subscribe((data: Task) => {
        this.task = data;
        this.item.task_id = this.task.id;
        this.template = _.find(this.route.parent.snapshot.data.templates, { name: this.task.template_name });
        this.taskIsResolvable = this.requestService.isResolvable(this.task, this.meta, this.template.allowed_resolver_usernames);
        if (this.task.resolution) {
          this.loadResolution(this.task.resolution).then(() => {
            resolve();
          }).catch(() => {
            reject();
          });
        } else {
          this.resolution = null;
          resolve();
        }
      }, (err: any) => {
        console.log(err);
        reject();
      }, () => {
        this.loaders.task = false;
      });
    });
  }

  loadResolution(resolutionId: string) {
    return new Promise((resolve, reject) => {
      this.loaders.resolution = true;
      this.api.resolution(resolutionId).subscribe((data) => {
        this.resolution = data;
        // this.generateSteps();
        resolve();
      }, (err: any) => {
        console.log(err);
        reject();
      }, () => {
        this.loaders.resolution = false;
      });
    });
  }
}
