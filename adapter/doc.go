/**
 * An adapter helps two incompatible interfaces to work together. This is the real world definition
 * for an adapter. Interfaces may be incompatible but the inner functionality should suit the need.
 * The Adapter design pattern allows otherwise incompatible classes to work together by converting
 * the interface of one class into an interface expected by the clients.
 *
 * There are two variations of the Adapter pattern: The class adapter implements the adaptee's
 * interface whereas the object adapter uses composition to contain the adaptee in the adapter
 * object. This example uses the object adapter approach.
 *
 * The Adapter FishingBoatAdapter converts the interface of the adaptee FishingBoat into a suitable
 * one expected by the client RowingBoat.
 *
 * The story of this implementation is this. Pirates are coming! we need a to flee! We have a
 * FishingBoat and our captain. We have no time to make up a new ship! we need to reuse this
 * FishingBoat. The captain needs a rowing boat which he can operate. The spec is in RowingBoat.
 * We will use the Adapter pattern to reuse FishingBoat.
 */
package main
