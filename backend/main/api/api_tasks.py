from rest_framework.decorators import api_view
from rest_framework.response import Response
from rest_framework.exceptions import ErrorDetail
from ..models import Task, Subtask, Column
from ..serializers.ser_task import TaskSerializer
from ..serializers.ser_subtask import SubtaskSerializer
from ..validation.val_auth import \
    authenticate, authorize, not_authenticated_response
from ..validation.val_column import validate_column_id
from ..validation.val_task import validate_task_id


@api_view(['GET', 'POST', 'PATCH', 'DELETE'])
def tasks(request):
    username = request.META.get('HTTP_AUTH_USER')
    token = request.META.get('HTTP_AUTH_TOKEN')

    user, authentication_response = authenticate(username, token)
    if authentication_response:
        return authentication_response

    # not in use – maintained for demonstration purposes
    if request.method == 'GET':
        column_id = request.query_params.get('column_id')

        validation_response = validate_column_id(column_id)
        if validation_response:
            return validation_response

        try:
            column = Column.objects.select_related(
                'board'
            ).prefetch_related(
                'task_set'
            ).get(id=column_id)
        except Column.DoesNotExist:
            return None, Response({
                'column_id': ErrorDetail(string='Column not found.',
                                         code='not_found')
            }, 404)

        if column.board.team_id != user.team.id:
            return not_authenticated_response

        return Response({
            'tasks': [
                {
                    'id': t['id'],
                    'order': t['order'],
                    'title': t['title'],
                    'description': t['description']
                } for t in column.task_set.all()
            ]
        }, 200)

    if request.method == 'POST':
        authorization_response = authorize(username)
        if authorization_response:
            return authorization_response

        column_id = request.data.get('column')
        validation_response = validate_column_id(column_id)
        if validation_response:
            return validation_response

        try:
            column = Column.objects.select_related(
                'board'
            ).prefetch_related(
                'task_set'
            ).get(id=column_id)
        except Column.DoesNotExist:
            return None, Response({
                'column_id': ErrorDetail(string='Column not found.',
                                         code='not_found')
            }, 404)

        if column.board.team_id != user.team_id:
            return not_authenticated_response

        for task in column.task_set.all():
            task.order += 1

        Task.objects.bulk_update(column.task_set.all(), ['order'])

        # save task
        task_serializer = TaskSerializer(
            data={'title': request.data.get('title'),
                  'description': request.data.get('description'),
                  'order': 0,
                  'column': request.data.get('column')}
        )
        if not task_serializer.is_valid():
            return Response(task_serializer.errors, 400)
        task = task_serializer.save()

        # save subtasks
        subtasks_data = [
            {
                'title': subtask,
                'order': i,
                'task': task.id
            } for i, subtask in enumerate(request.data.get('subtasks'))
        ]
        subtasks_serializer = SubtaskSerializer(data=subtasks_data, many=True)
        if not subtasks_serializer.is_valid():
            task.delete()
            return Response({
                'subtask': subtasks_serializer.errors
            }, 400)
        subtasks_serializer.save()

        return Response({
            'msg': 'Task creation successful.',
            'task_id': task.id
        }, 201)

    if request.method == 'PATCH':
        authorization_response = authorize(username)
        if authorization_response:
            return authorization_response

        task_id = request.query_params.get('id')
        task, validation_response = validate_task_id(task_id)

        if validation_response:
            return validation_response

        if task.column.board.team.id != user.team.id:
            return not_authenticated_response

        if 'title' in request.data.keys() and not request.data.get('title'):
            return Response({
                'title': ErrorDetail(string='Task title cannot be empty.',
                                     code='blank')
            }, 400)

        order = request.data.get('order')
        if 'order' in request.data.keys() and (order == '' or order is None):
            return Response({
                'order': ErrorDetail(string='Task order cannot be empty.',
                                     code='blank')
            }, 400)

        if 'column' in request.data.keys():
            column_id = request.data.get('column')
            _, validation_response = validate_column_id(column_id)
            if validation_response:
                return validation_response

        subtasks = request.data.pop('subtasks') \
            if 'subtasks' in request.data.keys() else None

        task_serializer = TaskSerializer(Task.objects.get(id=task_id),
                                         data=request.data,
                                         partial=True)
        if not task_serializer.is_valid():
            return Response(task_serializer.errors, 400)
        task = task_serializer.save()

        if subtasks:
            Subtask.objects.filter(task_id=task.id).delete()

            for subtask in subtasks:
                subtask_serializer = SubtaskSerializer(
                    data={'title': subtask['title'],
                          'order': subtask['order'],
                          'task': task.id,
                          'done': subtask['done']}
                )
                if not subtask_serializer.is_valid():
                    return Response({
                        'subtasks': subtask_serializer.errors
                    }, 400)
                subtask_serializer.save()

        return Response({
            'msg': 'Task update successful.',
            'id': task.id
        }, 200)

    if request.method == 'DELETE':
        authorization_response = authorize(username)
        if authorization_response:
            return authorization_response

        task_id = request.query_params.get('id')

        task, validation_response = validate_task_id(task_id)
        if validation_response:
            return validation_response

        if task.column.board.team.id != user.team.id:
            return not_authenticated_response

        task.delete()

        return Response({
            'msg': 'Task deleted successfully.',
            'id': task_id,
        })
