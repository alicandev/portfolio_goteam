# Generated by Django 3.1.7 on 2021-04-13 20:02

from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ('main', '0011_board_name'),
    ]

    operations = [
        migrations.AlterField(
            model_name='board',
            name='name',
            field=models.CharField(default='New Table', max_length=35),
        ),
    ]
