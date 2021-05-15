from rest_framework import serializers
from main.models import Subtask


class SubtaskSerializer(serializers.ModelSerializer):
    title = serializers.CharField(
        max_length=100,
        error_messages={
            'max_length':
                'Subtask titles cannot be longer than 50 characters.',
            'blank':
                'Subtask title cannot be empty.'
        }
    )

    class Meta:
        model = Subtask
        fields = '__all__'
