from __future__ import annotations

import copy
import traceback
from functools import reduce, wraps
from threading import Lock
from typing import (
    Optional,
    TypeAlias,
    cast,
    ParamSpec,
    TypeVar,
    Callable,
    Sequence,
)

from schema import Optional as Optional_, Schema, Use

from vula.sys_pyroute2 import Sys
from .common import raw, schemattrdict, yamlfile, yamlrepr_hl

from typing import Tuple, Any, TYPE_CHECKING

Trigger: TypeAlias = Tuple[str, Any]
TriggerResult: TypeAlias = Any

if TYPE_CHECKING:
    from vula.peer import Peer


class Result(yamlrepr_hl, schemattrdict):
    """
    A result object contains the results of an event, including the event,
    actions, writes, triggers, and trigger_results that resulted from
    the event.

    Although log replay is not yet implemented, the event engine is designed
    such that replaying the events from a log of result objects should produce
    an identical state and an identical series of result objects (except for
    the trigger_results, which depend on the system's actual configuration
    state, which exists outside of the state engine).
    """

    schema = Schema(
        {
            'event': Use(raw),
            'actions': Use(raw),
            'writes': Use(raw),
            Optional_('triggers'): Use(raw),
            Optional_('trigger_results'): Use(raw),
            Optional_('error'): object,
            Optional_('traceback'): str,
        },
    )

    default: dict[str, Any] | None = dict(
        event=[], actions=[], writes=[], triggers=[]
    )

    def __repr__(self) -> str:
        return "<Result\n  %s\n>" % (
            super(Result, self).__repr__().strip().replace("\n", "\n  "),
        )

    @property
    def ok(self) -> bool:
        return not self.error

    @property
    def error(self) -> Exception | None:
        return self.get('error')

    @property
    def triggers(self) -> list[Trigger]:
        return cast(list[Trigger], self.setdefault('triggers', []))

    @property
    def trigger_results(self) -> list[TriggerResult]:
        return cast(
            list[TriggerResult], self.setdefault('trigger_results', [])
        )

    @property
    def summary(self) -> str:
        if self.error:
            return "ERROR: %s" % (self.error,)
        else:
            return "OK: %s" % (" ".join(map(str, self.actions)))

    def add_triggers(self, **kw: Any) -> None:
        for name, args in kw.items():
            self.triggers.append((name, args))

    def run_triggers(self, target: object) -> Result:
        assert not self.trigger_results, "triggers should only be run once"
        for name, args in self.triggers:
            try:
                self.trigger_results.append(getattr(target, name)(*args))
            except Exception:
                self.trigger_results.append(traceback.format_exc())
        return self


ResultType: TypeAlias = Result
P = ParamSpec("P")
T = TypeVar("T")


class Engine(schemattrdict, yamlfile):
    """
    This is a transactional state engine. Subclasses implement rules in the
    form of events and actions.

    An event engine using this class may be thought of as a pure function which
    takes a current state and a new event, and produces a new state.

    Subclasses should define @Engine.event methods and @Engine.action methods.
    Events should call one or more actions. Actions may mutate the state via
    write operations, and may call other actions, and may register triggers
    which will be run after the transaction is committed.

    The three built-in write methods are SET, ADD, and REMOVE. These methods,
    called from actions during an event transaction, are the only places where
    the engine state is allowed to be modified.

    If there are exceptions during execution of an event and its actions,
    or if the state after all of the actions have been completed does not
    satisfy the schema, then none of the event's actions' write operations
    are applied.

    If the event does not have an error, after the new state is committed, the
    triggers are executed and their results are recorded in the event's
    result object. Triggers may modify state which exists outside of the state
    engine, and may also initiate new events.

    Calling an event will yield a Result object that contains a record of the
    event arguments and the resulting actions, writes, triggers, and trigger
    results, or contains the exception if one occurred.

    The state of the engine is only allowed to change through events which call
    actions, which in turn call write methods. The new state must depend solely
    on the old state and the event being processed. Inputs from the
    outside world, such as, for instance, the system time, need to be contained
    within an event to mutate the old state into the new state. Similarly,
    side effects should only happen in triggers which are run after the
    successful processing of an event.
    """

    Result = Result

    def __init__(self, *a: Any, **kw: Any) -> None:
        self._lock = Lock()
        self.result: Optional[Result] = None
        self.next_state: Optional[dict[str, Peer]] = None
        self.save: Callable[..., None] = lambda *a: None
        self.info_log: Callable[..., None] = lambda *a: None
        self.debug_log: Callable[..., None] = lambda *a: None
        self.trigger_target: Optional[Sys] = None
        super(Engine, self).__init__(*a, **kw)

    def record(self, result: ResultType) -> None:
        pass

    @staticmethod
    def event(method: Callable[P, T]) -> Callable[P, ResultType]:
        """
        Decorator for event methods
        """

        assert method.__name__.startswith('event_')
        name = method.__name__.split('_', 1)[1]

        @wraps(method)
        def _method(*args: P.args, **kwargs: P.kwargs) -> ResultType:
            self: Engine = cast(Engine, args[0])
            new_state = None
            res = self.Result(
                event=(name, *args),
                actions=[],
                writes=[],
                error=None,
            )
            error = None
            self._lock.acquire()
            try:
                self.next_state = copy.deepcopy(self._dict())
                self.result = res
                # run event method on a copy of our state
                method(*args, **kwargs)
                # confirm event produced a new valid state
                new_state = self.schema.validate(self.next_state)
                if raw(new_state) == raw(self):
                    self.debug_log("state unchanged")
                else:
                    # apply new state, cheating the ro_dict
                    dict.update(self, new_state)  # type: ignore
                    self._as_dict = None  # part of careful ro_dict cheating
                    self.save()
            except Exception as ex:
                error = [ex, traceback.format_exc()]
                data = res._dict()
                data.update(error=error[0], traceback=error[1], triggers=[])
                res = self.Result(**data)
            finally:
                self.result = None
                self.next_state = None
                self._lock.release()
            if self.trigger_target:
                res.run_triggers(self.trigger_target)
            self.record(res)
            self.debug_log(res)
            return res

        return _method

    @staticmethod
    def action(method: Callable[P, T]) -> Callable[P, dict[str, Peer]]:
        """
        Decorator for action methods.
        """
        assert method.__name__.startswith('action_')
        name = method.__name__.split('_', 1)[1]

        @wraps(method)
        def _method(*args: P.args, **kwargs: P.kwargs) -> dict[str, Peer]:
            self: Engine = cast(Engine, args[0])
            assert self.next_state, "can't run actions when not in an event"
            assert self.result
            self.result.actions.append((name,) + args)
            method(*args, **kwargs)
            return self.next_state

        return _method

    @staticmethod
    def write(
        method: Callable[[Engine, dict[str, Any], str, Any], T],
    ) -> Callable[[Engine, str | Sequence[str], Any], None]:
        """
        Decorator for write methods.

        Writes are where the state gets changed. They should be called from
        action methods, which should be called from event methods.

        The data model here is a bit weird and still not stable, but roughly
        speaking there are three write methods and they operate on these
        types:

            set:
                - any type
            add:
                - lists (treated as sorted sets)
                - sets
                - dicts of whatever, when new value is a dict
                - dicts of bools, when new value is not a dict
            remove:
                - lists (treated as sorted sets)
                - sets
                - dicts of whatever
        """

        assert method.__name__.startswith('_')
        name = method.__name__[1:]

        @wraps(method)
        def _method(
            self: Engine, path: str | Sequence[str], value: Any
        ) -> None:
            assert self.next_state, "can't run actions when not in an event"
            assert self.result
            self.result.writes.append((name, path, value))

            if type(path) is str:
                path = path.split('.')
            if len(path) == 1:
                target = self.next_state
                key = path[0]
            else:
                target = reduce(lambda a, b: a[b], path[:-1], self.next_state)
                key = path[-1]

            method(self, target, key, raw(value))

        return _method

    @write
    def _SET(self, target: dict[str, Any], key: str, value: Any) -> None:
        target[key] = value

    @write
    def _ADD(self, target: dict[str, Any], key: str, value: Any) -> None:
        if isinstance(target[key], (list, tuple)):
            target[key] = type(target[key])(
                item for item in target[key] if raw(item) != value
            ) + type(target[key])((value,))
        elif isinstance(target[key], (set, frozenset)):
            target[key] = type(target[key])(
                frozenset(raw(target[key])) | set([value])
            )
        elif isinstance(target[key], dict):
            if isinstance(value, dict):
                target[key].update(value)
            else:
                target[key].update({value: True})
        else:
            raise ValueError("Can't add type: %r" % type(target[key]))

    @write
    def _REMOVE(self, target: dict[str, Any], key: str, value: Any) -> None:
        if isinstance(target[key], (list, tuple)):
            target[key] = type(target[key])(
                item for item in raw(target[key]) if item != raw(value)
            )
        elif isinstance(target[key], (set, frozenset)):
            target[key] = type(target[key])(
                frozenset(raw(target[key])) - set([raw(value)])
            )
        elif isinstance(target[key], dict):
            del target[key][value]
        else:
            raise ValueError("Can't remove type: %r" % type(target[key]))


if __name__ == "__main__":
    import doctest

    doctest.testmod()
